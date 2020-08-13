package tidal_test

import (
	"testing"

	. "github.com/rotationalio/tidal"
	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	m, err := Open("testdata/0001_test_migration.sql")
	require.NoError(t, err)
	require.Equal(t, 1, m.Revision)
	require.Equal(t, "test migration", m.Name)
	require.False(t, m.Synchronized())

	pkg, err := m.Package()
	require.NoError(t, err)
	require.Equal(t, "foo", pkg)

	upsql, err := m.UpSQL()
	require.NoError(t, err)
	require.Contains(t, upsql, "CREATE TABLE IF NOT EXISTS groups")
	require.NotContains(t, upsql, "-- migrate: up")

	dnsql, err := m.DownSQL()
	require.NoError(t, err)
	require.Contains(t, dnsql, "DROP TABLE IF EXISTS users CASCADE;")
	require.NotContains(t, dnsql, "-- migrate: down")

	_, err = Open("testdata/foo.txt")
	require.EqualError(t, err, `could not parse "foo.txt" as a migration filename`)
}

func TestPredecessors(t *testing.T) {
	defer Reset()
	target := Migration{Revision: 3}

	// No migrations registered
	_, err := target.Predecessors()
	require.EqualError(t, err, "revision 3 was not registered")

	// Predecessors registered, but target not registered
	require.NoError(t, Register(Migration{Revision: 1}))
	require.NoError(t, Register(Migration{Revision: 2}))
	_, err = target.Predecessors()
	require.EqualError(t, err, "revision 3 was not registered")

	// Successors registered, but target not registered
	require.NoError(t, Register(Migration{Revision: 4}))
	require.NoError(t, Register(Migration{Revision: 5}))
	_, err = target.Predecessors()

	require.NoError(t, Register(target))
	n, err := target.Predecessors()
	require.NoError(t, err)
	require.Equal(t, 2, n)

	// No predecessors - at beginning of list
	target = Migration{Revision: 1}
	n, err = target.Predecessors()
	require.NoError(t, err)
	require.Equal(t, 0, n)
}

func TestSuccessors(t *testing.T) {
	defer Reset()
	target := Migration{Revision: 3}

	// No migrations registered
	_, err := target.Successors()
	require.EqualError(t, err, "revision 3 was not registered")

	// Predecessors registered, but target not registered
	require.NoError(t, Register(Migration{Revision: 1}))
	require.NoError(t, Register(Migration{Revision: 2}))
	_, err = target.Successors()
	require.EqualError(t, err, "revision 3 was not registered")

	// Successors registered, but target not registered
	require.NoError(t, Register(Migration{Revision: 4}))
	require.NoError(t, Register(Migration{Revision: 5}))
	_, err = target.Successors()

	require.NoError(t, Register(target))
	n, err := target.Successors()
	require.NoError(t, err)
	require.Equal(t, 2, n)

	// No successors - at end of list
	target = Migration{Revision: 5}
	n, err = target.Successors()
	require.NoError(t, err)
	require.Equal(t, 0, n)
}
