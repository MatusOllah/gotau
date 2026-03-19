package phonemizer

import (
	"iter"
	"regexp"
	"strings"

	"github.com/SladkyCitron/gotau/voicebank"
)

var _ Phonemizer = (*VCV)(nil)

// VCV is a simple vowel+consonant+vowel (VCV) [Phonemizer].
//
// It extracts the vowel from the end of the previous lyric and assumes lyrics
// contain explicit trailing vowel characters (e.g. romaji like "ka", "shi", "to").
// It does not support Japanese kana characters since the vowel
// cannot be determined from the last character. For kana-based Japanese voicebanks,
// use [JapaneseVCV].
//
// [VCV] emits candidates based on the following order:
//
//  1. prefix.map + lyric with VCV prefix (if [VCV.PrefixMap] and [ResolveConfig.PrevLyric] are present)
//  2. whitespace-trimmed lyric with VCV prefix (if [ResolveConfig.PrevLyric] is present)
//  3. raw lyric with VCV prefix (if [ResolveConfig.PrevLyric] is present)
type VCV struct {
	// PrefixMap contains the prefix.map rules for note-based prefix / suffix lookup.
	// Optional.
	PrefixMap voicebank.PrefixMap
}

var lastVowelRe = regexp.MustCompile(`(?i)[aeiouyあいうえおアイウエオ]$`)

// Resolve satisfies the [Phonemizer] interface.
func (p *VCV) Resolve(cfg ResolveConfig) iter.Seq[string] {
	return func(yield func(string) bool) {
		vowel := ""
		if cfg.PrevLyric != "" {
			vowel = lastVowelRe.FindString(cfg.PrevLyric)
		}

		vcvPrefix := "- "
		if vowel != "" {
			vcvPrefix = vowel + " "
		}

		vcvCombo := vcvPrefix + cfg.Lyric

		// prefix.map
		if entry, ok := p.PrefixMap[cfg.Note]; ok {
			if !yield(entry.Prefix + vcvCombo + entry.Suffix) {
				return
			}
		}

		// trimmed lyric
		if !yield(vcvPrefix + strings.TrimSpace(cfg.Lyric)) {
			return
		}

		// raw lyric
		// no need to check yield result; this is the final candidate
		yield(vcvCombo)
	}
}
