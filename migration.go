package tidal

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Used to parse a migration filename's components
var fnamere = regexp.MustCompile(`^(\d+)[_-]([\w\d_-]+)\.sql$`)

// Open a migration SQL file and parse it into a Migration object.
func Open(path string) (m Migration, err error) {
	filename := filepath.Base(path)
	if !fnamere.MatchString(filename) {
		return m, fmt.Errorf("could not parse %q as a migration filename", filename)
	}

	if m.Name, m.Revision, err = parseFilename(filename); err != nil {
		return m, err
	}

	// Now read the file and compress the contents into a descriptor
	var f *os.File
	if f, err = os.Open(path); err != nil {
		return m, err
	}
	defer f.Close()

	if m.descriptor, err = NewDescriptor(f, filename); err != nil {
		return m, err
	}

	return m, nil
}

// Migration defines how changes to the database are applied (up) or rolled back (down).
// Each migration is defined by two distinct pieces of SQL code, one for up and one for
// down, which are are parsed from a single SQL file, delimited by tidal-parseable
// comments. Migrations are generally compiled into a compressed descriptor format that
// can be included with application source code so that the migrations are compiled with
// the binary, rather than sourced from external files.
//
// Each migration can also include status information from the database (collected at
// runtime). This information defines the database's knowledge of the migration, e.g.
// has the migration been applied or not. Status information is stored in the database
// inside of a migrations table that is applied with Revision 0 (an application's first
// revision is Revision 1). The table is updated with migrate and sync commands.
//
// Migrations are identified by a unique revision number that specifies the sequence
// which migrations must be applied. For now that means that migrations can only be
// applied linearly (and not as a directed acyclic graph with multiple dependencies).
// Future work is required to create a migration DAG structure.
type Migration struct {
	Revision   int        // the unique id of the migration, prefix from the migration file
	Name       string     // the human readable name of the migration, suffix of the migration file
	Active     bool       // if the migration has been applied and is part of the active schema
	Applied    time.Time  // the timestamp the migration was applied
	Created    time.Time  // the timestamp the migration was added to the database
	descriptor Descriptor // contains the gzip compressed data to minimize compile time size
	dbsync     bool       // if the migration has been synchronized to the database
}

// Up applies the migration to the database. The migration creates a transaction that
// executes the SQL UP code as well as an update to the migrations table reflecting the
// change in state. Both of these SQL commands must be executed together without error
// otherwise the entire transaction is rolled back.
func (m *Migration) Up(conn *sql.DB) (err error) {
	var tx *sql.Tx
	if tx, err = conn.Begin(); err != nil {
		return fmt.Errorf("could not begin transaction to apply revision %d: %s", m.Revision, err)
	}

	defer func() {
		// Recover from panic, rolling back transaction, then re-throw panic
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			// Rollback the transaction, but don't get the rollback error since the
			// error is already non nil, and that's what we want to return
			tx.Rollback()
		} else {
			// Success, commit! Store any commit errors to return if necessary
			err = tx.Commit()
		}
	}()

	// Execute up transaction
	err = m.upTx(tx)
	return err
}

func (m *Migration) upTx(tx *sql.Tx) (err error) {
	var sql string
	if sql, err = m.UpSQL(); err != nil {
		return fmt.Errorf("could not parse revision %d up sql: %s", m.Revision, err)
	}

	if _, err = tx.Exec(sql); err != nil {
		return fmt.Errorf("could not exec revision %d up: %s", m.Revision, err)
	}

	// If this is an application migration, update the migrations status table
	if m.Revision > 0 {
		sql := "UPDATE migrations SET active=$1, applied=$2 WHERE revision=$3"
		if _, err = tx.Exec(sql, true, time.Now().UTC(), m.Revision); err != nil {
			return fmt.Errorf("could not update migration status of revision %d: %s", m.Revision, err)
		}
	}

	return nil
}

// UpSQL returns the sql statement defined for applying the migration to the specific
// revision. This requires parsing the underlying descriptor correctly.
func (m *Migration) UpSQL() (string, error) {
	return m.descriptor.Up()
}

// Down rolls back the migration from the database. The migration creates a transaction
// that executes the SQL DOWN code as well as an update to the migrations table reflecting
// the change in state. Both of these SQL commands must be executed together without
// error, otherwise the entire transaction is rolled back.
func (m *Migration) Down(conn *sql.DB) (err error) {
	var tx *sql.Tx
	if tx, err = conn.Begin(); err != nil {
		return fmt.Errorf("could not begin transaction to rollback revision %d: %s", m.Revision, err)
	}

	defer func() {
		// Recover from panic, rolling back transaction, then re-throw panic
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			// Rollback the transaction, but don't get the rollback error since the
			// error is already non nil, and that's what we want to return
			tx.Rollback()
		} else {
			// Success, commit! Store any commit errors to return if necessary
			err = tx.Commit()
		}
	}()

	// Execute down transaction
	err = m.downTx(tx)
	return err
}

func (m *Migration) downTx(tx *sql.Tx) (err error) {
	var sql string
	if sql, err = m.DownSQL(); err != nil {
		return fmt.Errorf("could not parse revision %d down sql: %s", m.Revision, err)
	}

	if _, err = tx.Exec(sql); err != nil {
		return fmt.Errorf("could not exec revision %d down: %s", m.Revision, err)
	}

	// If this is an application migration, update the migrations status table
	if m.Revision > 0 {
		sql := "UPDATE migrations SET active=$1, applied=NULL WHERE revision=$3"
		if _, err = tx.Exec(sql, false, m.Revision); err != nil {
			return fmt.Errorf("could not update migration status of revision %d: %s", m.Revision, err)
		}
	}

	return nil
}

// DownSQL returns the sql statement defined for rolling back the migration to a state
// before this specific revision. This requires parsing the underlying descriptor correctly.
func (m *Migration) DownSQL() (string, error) {
	return m.descriptor.Down()
}

// Package returns the parsed package directive from the descriptor if it has one.
func (m *Migration) Package() (string, error) {
	return m.descriptor.Package()
}

// Synchronized returns true if the migration state has been synchronized with the database.
func (m *Migration) Synchronized() bool {
	return m.dbsync
}

// Predecessors returns the number of migrations before this migration.
func (m *Migration) Predecessors() (n int, err error) {
	if len(migrations) == 0 {
		return 0, fmt.Errorf("revision %d was not registered", m.Revision)
	}

	for _, o := range migrations {
		if m.Revision == o.Revision {
			break
		}
		if o.Revision > m.Revision {
			return 0, fmt.Errorf("revision %d was not registered", m.Revision)
		}
		n++
	}

	if n == len(migrations) && migrations[n-1].Revision != m.Revision {
		return 0, fmt.Errorf("revision %d was not registered", m.Revision)
	}
	return n, nil
}

// Successors returns the number of migrations after this migration.
func (m *Migration) Successors() (n int, err error) {
	i := sort.Search(len(migrations), func(i int) bool {
		return m.Revision <= migrations[i].Revision
	})

	if i < len(migrations) && migrations[i].Revision == m.Revision {
		if i+1 == len(migrations) {
			return 0, nil
		}
		return len(migrations[i+1:]), nil
	}
	return 0, fmt.Errorf("revision %d was not registered", m.Revision)
}

// helper function parse a filename or path into Migration metadata
func parseFilename(filename string) (name string, revision int, err error) {
	groups := fnamere.FindStringSubmatch(filename)
	name = strings.Replace(groups[2], "_", " ", -1)
	if revision, err = strconv.Atoi(groups[1]); err != nil {
		return "", 0, fmt.Errorf("could not parse %q to revision number: %s", groups[1], err)
	}
	return name, revision, nil
}
