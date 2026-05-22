// Package phonemizer implements the lyric to alias resolution and phonemization logic
// and provides multiple ready-to-use implementations.
package phonemizer

import (
	"iter"

	"gitlab.com/gomidi/midi/v2"
)

// Note represents a note that is used as input for the phonemizer.
type Note struct {
	// Position is the position of the note in MIDI ticks.
	Position int

	// Duration is the duration of the note in MIDI ticks.
	Duration int

	// Lyric is the lyric of the note. It can be either latin characters or kana.
	Lyric string

	// Note is the MIDI note number.
	Note midi.Note
}

// Phoneme represents a phoneme.
type Phoneme struct {
	// Index is the index of the note this phoneme points to.
	Index int

	// Candidates is a list of candidate oto aliases for this phoneme. The first candidate should be the most likely one.
	Candidates []string

	// Error is an error that occurred during phonemization.
	Error error
}

// Phonemizer is the interface that is implemented by phonemizers and wraps the basic Phonemize method.
// It resolves a list of input slur notes and their context (previous and next note) into phonemes
// with oto alias candidates and returns an lazy iterator over them evaluated in order.
//
// Phonemizer encapsulates the logic and rules for converting a lyric into a sequence
// of phonemes suitable for oto lookup and final synthesis.
// This allows different voicebanks and spoken languages support various
// phonemization schemes (e.g. CV, VCV) and the synthesis engine to use
// any phonemization scheme that is supported by the voicebank or language.
//
// # Input notes and context
//
// The input notes are a slice of notes that are slurred (tied) together and should be phonemized together as one unit.
// The first note in the slice is the main note that determines the lyric and pitch, while the rest are extended
// slur notes (i.e. "+") that only contribute to the phonemization of the main note.
//
// The previous and next note are adjacent notes to the main note and are used for context-aware phonemization (e.g. VCV, CVVC).
// They can be nil if there is no previous or next note.
type Phonemizer interface {
	Phonemize(notes []Note, prev *Note, next *Note) iter.Seq[Phoneme]
}
