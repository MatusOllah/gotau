package voicebank_test

import (
	"strings"
	"testing"

	"github.com/MatusOllah/gotau/voicebank"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/encoding/japanese"
)

func TestDecodeOto(t *testing.T) {
	src := `; example oto.ini for parser tests
a.wav=a,0,120,-39,80,20
shi.wav=し,10,90,-200,70,25
ka.wav=か,50,100,-300,90,30
あ.wav=あ,39,110,-250,85,40
empty_values.wav=a,,,,,
`
	want := voicebank.Oto{
		{"a.wav", "a", 0, 120, -39, 80, 20},
		{"shi.wav", "し", 10, 90, -200, 70, 25},
		{"ka.wav", "か", 50, 100, -300, 90, 30},
		{"あ.wav", "あ", 39, 110, -250, 85, 40},
		{"empty_values.wav", "a", 0, 0, 0, 0, 0},
	}

	oto, err := voicebank.DecodeOto(strings.NewReader(src), voicebank.OtoWithComment(';'))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, want, oto)
}

func TestDecodeOtoShiftJIS(t *testing.T) {
	src := "shi.wav=\x82\xb5,10,90,-200,70,25\nka.wav=\x82\xa9,50,100,-300,90,30\n\x82\xa0.wav=\x82\xa0,39,110,-250,85,40\n"

	want := voicebank.Oto{
		{"shi.wav", "し", 10, 90, -200, 70, 25},
		{"ka.wav", "か", 50, 100, -300, 90, 30},
		{"あ.wav", "あ", 39, 110, -250, 85, 40},
	}

	oto, err := voicebank.DecodeOto(strings.NewReader(src), voicebank.OtoWithEncoding(japanese.ShiftJIS))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, want, oto)
}

func TestDecodeOtoInvalid(t *testing.T) {
	src := "invalid_entry.wav=a,\n"

	oto, err := voicebank.DecodeOto(strings.NewReader(src))
	assert.Error(t, err)
	assert.Empty(t, oto)
}

func TestOto_Get(t *testing.T) {
	oto := voicebank.Oto{
		{"shi.wav", "し", 10, 90, -200, 70, 25},
		{"ka.wav", "か", 50, 100, -300, 90, 30},
		{"あ.wav", "あ", 39, 110, -250, 85, 40},
	}

	want := voicebank.OtoEntry{"ka.wav", "か", 50, 100, -300, 90, 30}

	entry, ok := oto.Get("か")
	assert.True(t, ok)
	assert.Equal(t, want, entry)

	entry, ok = oto.Get("nonexistent")
	assert.False(t, ok)
	assert.Empty(t, entry)
}
