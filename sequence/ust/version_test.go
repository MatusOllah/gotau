package ust_test

import (
	"fmt"
	"testing"

	"github.com/MatusOllah/gotau/sequence/ust"
	"github.com/stretchr/testify/assert"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name          string
		s             string
		expectedVer   ust.Version
		expectedError error
	}{
		{name: "Version1_2", s: "UST Version1.2", expectedVer: ust.Version1_2, expectedError: nil},
		{name: "Version2_0", s: "UST Version2.0", expectedVer: ust.Version2_0, expectedError: nil},
		{name: "Invalid", s: "UST Version3.9", expectedVer: 0, expectedError: fmt.Errorf("invalid version string: UST Version3.9")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ver, err := ust.ParseVersion(test.s)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedVer, ver)
		})
	}
}

func TestVersion_String(t *testing.T) {
	assert.Equal(t, "1.2", ust.Version1_2.String())
	assert.Equal(t, "2.0", ust.Version2_0.String())
}

func TestVersion_RawString(t *testing.T) {
	assert.Equal(t, "UST Version1.2", ust.Version1_2.RawString())
	assert.Equal(t, "UST Version2.0", ust.Version2_0.RawString())
}
