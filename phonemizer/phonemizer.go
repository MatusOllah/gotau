// Package phonemizer implements the lyric to alias resolution and phonemization logic.
package phonemizer

import (
	"iter"

	"gitlab.com/gomidi/midi/v2"
)

// Phonemizer is the interface that is implemented by phonemizers and wraps the basic Resolve method.
// It resolves a lyric into phoneme alias candidates and returns an iterator over them.
//
// Phonemizer encapsulates the logic and rules for converting a lyric into a sequence
// of candidate phoneme aliases suitable for oto lookup and final voice synthesis.
// This allows different voicebanks support various phonemization schemes (e.g. CV, VCV)
// and the synthesis engine to use any phonemization scheme that is supported by the voicebank.
type Phonemizer interface {
	Resolve(cfg ResolveConfig) iter.Seq[string]
}

// ResolveConfig represents the configuration for passing into [Phonemizer.Resolve].
type ResolveConfig struct {
	// Lyric is the main lyric.
	Lyric string

	// PrevLyric is the previous lyric.
	PrevLyric string

	// Note is the MIDI note. It's used for prefix.map lookup.
	Note midi.Note
}
