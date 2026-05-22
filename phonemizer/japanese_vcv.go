package phonemizer

import (
	"iter"
	"slices"

	"golang.org/x/text/unicode/norm"
)

// loosely based on OpenUtau
// https://github.com/stakira/OpenUtau/blob/master/OpenUtau.Plugin.Builtin/JapaneseVCVPhonemizer.cs

var _ Phonemizer = (*JapaneseVCV)(nil)

// JapaneseVCV is a Japanese vowel+consonant+vowel (VCV) [Phonemizer].
//
// It extracts the vowel from the final kana of the previous lyric.
type JapaneseVCV struct{}

// Phonemize satisfies the [Phonemizer] interface.
func (p *JapaneseVCV) Phonemize(notes []Note, prev *Note, _ *Note) iter.Seq[Phoneme] {
	return func(yield func(Phoneme) bool) {
		combos := make([]string, 0, 4)

		note := notes[0]
		lyric := norm.NFC.String(note.Lyric)

		//TODO: phonetic hints??

		// aliases for when there's no previous note
		combos = append(combos, "- "+lyric, lyric)

		// previous note
		if prev != nil {
			prevLyric := norm.NFC.String(prev.Lyric)
			vowel := kanaTailVowel(prevLyric)
			combos = combos[:0]
			combos = append(combos,
				string(vowel)+" "+lyric,
				"* "+lyric,
				lyric,
				"- "+lyric,
			)
		}

		yield(Phoneme{Index: 0, Candidates: slices.Compact(combos)})
	}
}
