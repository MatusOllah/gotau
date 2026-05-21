package phonemizer

import "iter"

var _ Phonemizer = (*Default)(nil)

// Default is the simplest [Phonemizer] possible.
// It simply passes the lyric as the phoneme without any other lyric pre-processing.
type Default struct{}

// Phonemize satisfies the [Phonemizer] interface.
func (p *Default) Phonemize(notes []Note, _ *Note, _ *Note) iter.Seq[Phoneme] {
	return func(yield func(Phoneme) bool) {
		yield(Phoneme{Index: 0, Candidates: []string{notes[0].Lyric}})
	}
}
