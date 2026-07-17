// Package concat defines the Concatenator interface and provides multiple implementations.
//
// "Concatenator" refers to the UTAU wavtool.
// It is responsible for concatenating a resampled note with the previous note's tail
// to create a smooth transition between notes.
package concat

import (
	"github.com/SladkyCitron/gotau/umath"
	"github.com/SladkyCitron/resona/afmt"
)

// Concatenator concatenates a resampled note with the tail to create a smooth transition between notes.
type Concatenator interface {
	// Concatenate appends the note to the tail using the provided configuration and returns the resulting audio data.
	Concatenate(tail []float32, note []float32, cfg ConcatenateConfig) ([]float32, error)
}

// ConcatenateConfig represents the configuration for passing into [Concatenator.Concatenate].
type ConcatenateConfig struct {
	// Offset is the offset time in milliseconds.
	Offset float64

	// Duration is the duration of the note in MIDI ticks.
	// It's used for calculating the note length.
	Duration int

	// Tempo is the tempo in beats per minute (BPM).
	// // It's used for calculating the note length.
	Tempo float64

	// Resolution is the number of MIDI ticks per quarter note (TPQN).
	// It's used for calculating the note length.
	Resolution int

	// LengthDelta is the correction value for the note length calculation.
	// It's used for adjusting the note length calculation.
	LengthDelta float64

	// Overlap is the overlap time in milliseconds (from oto).
	Overlap float64

	// Envelope is the envelope curve to apply to the concatenated note.
	// Usually the curve only has 5 points and linear interpolation.
	Envelope umath.Envelope

	// AudioFormat is the audio format of the input and output audio data.
	AudioFormat afmt.Format
}
