/*
Package tidal provides a mechanism to define and manage database schema migrations using
SQL files that both specify the up (apply) and down (rollback) actions to ensure consistent
changes in the database schema as application versions change. Tidal provides a CLI tool
to generate these files into descriptors, directly adding them to your application source
code so they can be compiled into the binary. The tidal package provides utilities for
managing the state of the database with respect to the migrations, even across different
binaries and application versions.
*/
package tidal

import (
	"errors"
	"fmt"
	"sort"
)

// Contains all migrations that have been registered by the application as well as the
// initial migration for creating the migrations database. Most migrations are added to
// this data structure using the generated code registration functions. The tidal
// package then manages the database with respect to these migrations.
var migrations []Migration

// Register a migration to be managed by tidal. Note that although migrations can be
// directly applied using the Migration interface, they must be registered in order to
// preserve dependency order. It is highly recommended to register migrations and to
// use the tidal migration interface rather than managing migrations manually.
func Register(m Migration) (err error) {
	// Maintain the migrations array sorted by revision id
	i := sort.Search(len(migrations), func(i int) bool { return migrations[i].Revision >= m.Revision })
	if i < len(migrations) && migrations[i].Revision == m.Revision {
		return fmt.Errorf("cannot register migration with revision %d: revision already exists", m.Revision)
	}

	// Insort the migration into the migrations array
	migrations = append(migrations, Migration{})
	copy(migrations[i+1:], migrations[i:])
	migrations[i] = m
	return nil
}

// RegisterDescriptor creates a Migration from descriptor data and registers it.
func RegisterDescriptor(data []byte) (err error) {
	m := Migration{
		descriptor: Descriptor(data),
	}

	var filename string
	if filename, _, err = m.descriptor.Info(); err != nil {
		return err
	}

	if filename == "" {
		return errors.New("descriptor data does not contain required header information")
	}

	if m.Name, m.Revision, err = parseFilename(filename); err != nil {
		return err
	}

	return Register(m)
}

// Reset removes all registered migrations. Primarily used for testing.
func Reset() (err error) {
	migrations = make([]Migration, 0)
	return nil
}

// ByRevision implements sort.Interface for []Migration based on the Revision field.
type ByRevision []Migration

func (a ByRevision) Len() int           { return len(a) }
func (a ByRevision) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRevision) Less(i, j int) bool { return a[i].Revision < a[j].Revision }
