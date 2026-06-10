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

// Concatenator concatenates a resampled note with the previous note's tail to create a smooth transition between notes.
type Concatenator interface {
	// Concatenate concatenates the given note with the tail of the previous note using the provided configuration.
	Concatenate(tail []float32, note []float32, cfg ConcatenateConfig) ([]float32, error)
}

// ConcatenateConfig represents the configuration for passing into [Concatenator.Concatenate].
type ConcatenateConfig struct {
	// Offset is the offset time in milliseconds (from oto).
	Offset float64

	// Length is the desired length of the final output in milliseconds.
	Length float64

	// Overlap is the overlap time in milliseconds (from oto).
	Overlap float64

	// Envelope is the envelope curve to apply to the concatenated note.
	// Usually the curve only has 5 points and linear interpolation.
	Envelope umath.Curve

	// AudioFormat is the audio format of the input and output audio data.
	AudioFormat afmt.Format
}
