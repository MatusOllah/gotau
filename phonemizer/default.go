package phonemizer

import "iter"

var _ Phonemizer = (*Default)(nil)

// Default is the simplest [Phonemizer] possible.
// It simply passes the lyric as the phoneme without any prefix.map lookup
// or other lyric pre-processing.
type Default struct{}

// Resolve satisfies the [Phonemizer] interface.
func (p *Default) Resolve(cfg ResolveConfig) iter.Seq[string] {
	return func(yield func(string) bool) {
		yield(cfg.Lyric)
	}
}
