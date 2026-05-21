package phonemizer

import (
	"iter"
	"slices"
	"strings"

	"github.com/SladkyCitron/gotau/voicebank"
)

var _ Phonemizer = (*CV)(nil)

// CV is a simple consonant+vowel (CV) [Phonemizer].
type CV struct {
	// PrefixMap contains the prefix.map rules for note-based prefix / suffix lookup.
	// Optional.
	PrefixMap voicebank.PrefixMap
}

// Phonemize satisfies the [Phonemizer] interface.
func (p *CV) Phonemize(notes []Note, _ *Note, _ *Note) iter.Seq[Phoneme] {
	return func(yield func(Phoneme) bool) {
		combos := make([]string, 0, 4)

		// prefix.map
		if p.PrefixMap != nil {
			if entry, ok := p.PrefixMap[notes[0].Note]; ok {
				combos = append(combos, entry.Prefix+strings.TrimSpace(notes[0].Lyric)+entry.Suffix)
				combos = append(combos, entry.Prefix+notes[0].Lyric+entry.Suffix)
			}
		}

		// trimmed lyric
		combos = append(combos, strings.TrimSpace(notes[0].Lyric))

		// raw lyric
		combos = append(combos, notes[0].Lyric)

		yield(Phoneme{Index: 0, Candidates: slices.Compact(combos)})
	}
}
