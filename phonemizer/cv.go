package phonemizer

import (
	"iter"
	"strings"

	"github.com/SladkyCitron/gotau/voicebank"
)

var _ Phonemizer = (*CV)(nil)

// CV is a simple consonant+vowel (CV) [Phonemizer].
//
// It emits candidates based on the following order:
//
//  1. prefix.map combo (if [CV.PrefixMap] is present)
//  2. whitespace-trimmed lyric
//  3. raw lyric
type CV struct {
	// PrefixMap contains the prefix.map rules for note-based prefix / suffix lookup.
	// Optional.
	PrefixMap voicebank.PrefixMap
}

// Resolve satisfies the [Phonemizer] interface.
func (p *CV) Resolve(cfg ResolveConfig) iter.Seq[string] {
	return func(yield func(string) bool) {
		// prefix.map
		if entry, ok := p.PrefixMap[cfg.Note]; ok {
			if !yield(entry.Prefix + cfg.Lyric + entry.Suffix) {
				return
			}
		}

		// trimmed lyric
		if !yield(strings.TrimSpace(cfg.Lyric)) {
			return
		}

		// raw lyric
		// no need to check yield result; this is the final candidate
		yield(cfg.Lyric)
	}
}
