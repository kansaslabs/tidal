package tidal

import "fmt"

// Semantic versioning components for easy package comparsion.
const (
	VersionMajor = 1
	VersionMinor = 0
	VersionPatch = 0
)

// Version returns a string with the semvar version of the current build.
func Version() string {
	if VersionPatch > 0 {
		return fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
	}
	return fmt.Sprintf("%d.%d", VersionMajor, VersionMinor)
}
