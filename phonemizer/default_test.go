package phonemizer_test

import (
	"slices"
	"testing"

	"github.com/SladkyCitron/gotau/phonemizer"
	"github.com/stretchr/testify/assert"
)

func TestDefault(t *testing.T) {
	want := "a"

	p := &phonemizer.Default{}
	got := slices.Collect(p.Phonemize([]phonemizer.Note{{Lyric: want}}, nil, nil))

	wantSlice := []phonemizer.Phoneme{{Index: 0, Candidates: []string{want}}}
	assert.Equal(t, wantSlice, got)
}
