package phonemizer_test

import (
	"slices"
	"testing"

	"github.com/SladkyCitron/gotau/phonemizer"
	"github.com/stretchr/testify/assert"
)

func TestCV(t *testing.T) {
	p := &phonemizer.CV{}
	got := slices.Collect(p.Phonemize([]phonemizer.Note{{Lyric: " a "}}, nil, nil))

	want := []phonemizer.Phoneme{{Index: 0, Candidates: []string{
		"a",   // trimmed
		" a ", // raw
	}}}

	assert.Equal(t, want, got)
}
