package tidal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeterminePackage(t *testing.T) {
	t.Skip("requires descriptors to be tested")
	migrations := []Migration{
		{Revision: 1},
		{Revision: 2},
	}

	pkg, err := determinePackage(migrations, "")
	require.NoError(t, err)
	require.Equal(t, "foo", pkg)

}
