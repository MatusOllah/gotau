package phonemizer_test

import (
	"slices"
	"testing"

	"github.com/SladkyCitron/gotau/phonemizer"
	"github.com/SladkyCitron/gotau/voicebank"
	"github.com/stretchr/testify/assert"
)

func TestCV(t *testing.T) {
	p := &phonemizer.CV{PrefixMap: voicebank.PrefixMap{60: voicebank.Prefix{"pre", "suf"}}}
	got := slices.Collect(p.Phonemize([]phonemizer.Note{{Lyric: " a ", Note: 60}}, nil, nil))

	want := []phonemizer.Phoneme{{Index: 0, Candidates: []string{
		"preasuf",   // prefix.map trimmed
		"pre a suf", // prefix.map
		"a",         // trimmed
		" a ",       // raw
	}}}

	assert.Equal(t, want, got)
}
