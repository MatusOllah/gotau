package sequence

import (
	"time"

	"github.com/SladkyCitron/gotau/internal/timeutil"
	"gitlab.com/gomidi/midi/v2"
)

// Sequence represents a voice synth note sequence. It holds metadata and a list of notes.
type Sequence struct {
	// Metadata is the metadata.
	Metadata Metadata

	// Notes is the list of notes. It should be sorted by position (ascending).
	Notes []Note
}

// Metadata represents the metadata of a sequence.
type Metadata struct {
	// Name is the human-readable name of the sequence (e.g. project name, song name).
	Name string

	// VoicebankPath is the path to the voicebank used by the sequence.
	VoicebankPath string

	// OutputPath is the path to the final output file (e.g. wav file).
	OutputPath string

	// Resolution is the number of MIDI ticks per quarter note (TPQN).
	Resolution int

	// Tempo is the tempo of the sequence in beats per minute (BPM).
	Tempo float32
}

// time signature is always 4/4 (for now)

// Note represents a single musical note in the sequence.
type Note struct {
	// Position is the position of the note in MIDI ticks.
	Position int

	// Duration is the duration of the note in MIDI ticks.
	Duration int

	// Lyric is the lyric of the note. It can be either latin characters or kana.
	Lyric string

	// Note is the MIDI note number.
	Note midi.Note

	// Intensity is the loudness or intensity of the note (0.0 to 1.0).
	Intensity float32

	// PreUtterance is the duration (in milliseconds) before note to start playback (in OTO). If it's omitted, falls back to OTO defaults.
	PreUtterance *float32

	// VoiceOverlap is the amount of overlap into the previous note. If it's omitted, falls back to OTO defaults.
	VoiceOverlap *float32

	// StartPoint is the time where to begin sampling inside the audio file (in milliseconds). If it's omitted, falls back to OTO defaults.
	StartPoint *float32

	// Envelope is the volume envelope curve. It should have at most 5 points.
	Envelope Curve

	// PitchBend is the pitch bend curve.
	PitchBend Curve
}

// CurveInterpolation is the type of interpolation used between curve points.
type CurveInterpolation uint8

const (
	CurveInterpolationLinear CurveInterpolation = iota
	CurveInterpolationSine
	CurveInterpolationRigid
	CurveInterpolationJump
)

// Curve represents a curve. It consists of a list of curve points and the interpolation type between them.
type Curve []CurvePoint

// CurvePoint represents a single point on a curve.
type CurvePoint struct {
	// X is the position in MIDI ticks.
	X int

	// Y is the value.
	Y float32

	// Interp is the interpolation type to the next point. If it's the last point, it is ignored.
	Interp CurveInterpolation
}

// Sequencer is the interface for something that can produce a [Sequence].
type Sequencer interface {
	// Sequence returns the [Sequence].
	Sequence() Sequence
}

// Len returns the sequence's length in MIDI ticks.
func (s Sequence) Len() int {
	len := 0
	for _, note := range s.Notes {
		len += note.Duration
	}
	return len
}

// Duration returns the sequence's length as a [time.Duration].
func (s Sequence) Duration() time.Duration {
	return timeutil.TicksToDuration(s.Len(), s.Metadata.Resolution, float64(s.Metadata.Tempo))
}
