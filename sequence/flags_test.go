package sequence_test

import (
	"testing"

	"github.com/SladkyCitron/gotau/sequence"
	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	want := sequence.Flags{
		"g": -5,
		"B": 10,
		"N": sequence.OptionFlagValue,
		"H": 0,
		"P": 86,
	}

	got, err := sequence.ParseFlags("g-5B10NH0P86")
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestParseFlags_Empty(t *testing.T) {
	want := sequence.Flags{}

	got, err := sequence.ParseFlags("")
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestParseFlags_InvalidCharacter(t *testing.T) {
	_, err := sequence.ParseFlags("g-5/B10")
	assert.ErrorContains(t, err, `sequence ParseFlags: invalid character in flags: '/'`)
}

func TestFlags_String(t *testing.T) {
	flags := sequence.Flags{
		"g": -5,
		"B": 10,
	}
	assert.Equal(t, "B10g-5", flags.String())
}

func TestFlags_String_Empty(t *testing.T) {
	flags := sequence.Flags{}
	assert.Equal(t, "", flags.String())
}

func TestFlags_Has(t *testing.T) {
	flags := sequence.Flags{
		"g": -5,
		"B": 10,
	}
	assert.Equal(t, flags.Has("g"), true)
	assert.Equal(t, flags.Has("B"), true)
	assert.Equal(t, flags.Has("H"), false)
}
