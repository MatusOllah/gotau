package resample

import (
	"github.com/SladkyCitron/gotau/sequence"
	"github.com/SladkyCitron/resona/afmt"
	"github.com/SladkyCitron/resona/aio"
	"gitlab.com/gomidi/midi/v2"
)

// Resampler is the interface for resamplers. It is responsible for resampling notes based on
// the provided configuration (pitch, velocity, oto, pitch bend, etc.).
//
// It should also be deterministic, meaning that it should return the same output for the same input and configuration.
type Resampler interface {
	Resample(in aio.SampleReader, cfg ResampleConfig) (aio.SampleReader, error)
}

// ResampleConfig represents the configuration for passing into [Resampler.Resample].
type ResampleConfig struct {
	// Pitch is the MIDI note number to resample to.
	Pitch midi.Note

	// Velocity is the velocity. It is a value between 0 and 1, where 1 is the maximum velocity.
	Velocity float64

	// Flags is a string of flags to pass to the resampler. These can be resampler-specific.
	Flags string

	// Offset is the offset time in milliseconds (from oto).
	Offset float64

	// Length is the desired length of the final resampled note.
	Length float64

	// Consonant is the consonant duration in milliseconds (from oto).
	Consonant float64

	// Cutoff is the cutoff time in milliseconds (from oto).
	Cutoff float64

	// Volume is the volume of the note.
	Volume float64

	// Modulation is the modulation (vibrato) depth.
	Modulation float64

	// Tempo is the tempo in BPM.
	Tempo float64

	// Resolution is the timing resolution in ticks per quarter note (TPQN).
	// It is only used for sampling the pitch bend curve.
	Resolution int

	// PitchBend is the pitch bend curve. It is a curve that maps time (in MIDI ticks) to pitch (in semitones).
	PitchBend sequence.Curve

	// AudioFormat is the audio format of the input and output audio data.
	AudioFormat afmt.Format
}
