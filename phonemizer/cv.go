package phonemizer

import (
	"iter"
	"slices"
	"strings"
)

var _ Phonemizer = (*CV)(nil)

// CV is a simple consonant+vowel (CV) [Phonemizer].
type CV struct{}

// Phonemize satisfies the [Phonemizer] interface.
func (p *CV) Phonemize(notes []Note, _ *Note, _ *Note) iter.Seq[Phoneme] {
	return func(yield func(Phoneme) bool) {
		combos := make([]string, 0, 2)

		// trimmed lyric
		combos = append(combos, strings.TrimSpace(notes[0].Lyric))

		// raw lyric
		combos = append(combos, notes[0].Lyric)

		yield(Phoneme{Index: 0, Candidates: slices.Compact(combos)})
	}
}
