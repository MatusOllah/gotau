package ust

import "fmt"

// Version represents a UST file version.
type Version int

const (
	Version1_2 Version = iota
	Version2_0
)

// String returns a string representation of the version.
func (ver Version) String() string {
	switch ver {
	case Version1_2:
		return "1.2"
	case Version2_0:
		return "2.0"
	default:
		panic("invalid version")
	}
}

// RawString returns a raw UST string representation of the version.
func (ver Version) RawString() string {
	return "UST Version" + ver.String()
}

// ParseVersion parses a version string.
func ParseVersion(s string) (Version, error) {
	switch s {
	case "UST Version1.2":
		return Version1_2, nil
	case "UST Version2.0":
		return Version2_0, nil
	default:
		return 0, fmt.Errorf("invalid version string: %s", s)
	}
}
