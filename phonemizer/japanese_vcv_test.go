package phonemizer_test

import (
	"slices"
	"testing"

	"github.com/SladkyCitron/gotau/phonemizer"
	"github.com/stretchr/testify/assert"
)

func TestJapaneseVCV(t *testing.T) {
	p := &phonemizer.JapaneseVCV{}
	got := slices.Collect(p.Phonemize([]phonemizer.Note{{Lyric: "か "}}, nil, nil))

	want := []phonemizer.Phoneme{{Index: 0, Candidates: []string{
		"- か ", // no prev
		"か ",   // lyric
	}}}

	assert.Equal(t, want, got)
}

func TestJapaneseVCV_PrevLyric(t *testing.T) {
	p := &phonemizer.JapaneseVCV{}
	got := slices.Collect(p.Phonemize([]phonemizer.Note{{Lyric: "ク"}}, &phonemizer.Note{Lyric: "ミ"}, nil))

	want := []phonemizer.Phoneme{{Index: 0, Candidates: []string{
		"i ク",
		"* ク",
		"ク",
		"- ク",
	}}}

	assert.Equal(t, want, got)
}
