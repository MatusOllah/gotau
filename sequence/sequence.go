package sequence

import (
	"github.com/SladkyCitron/gotau/umath"
	"gitlab.com/gomidi/midi/v2"
)

type Sequence struct {
	Metadata Metadata
	Notes    []Note
}

type Metadata struct {
	Name          string
	VoicebankPath string
	OutputPath    string
	Resolution    int
	Tempo         float32
}

// time signature is always 4/4 (for now)

type Note struct {
	Position     int
	Duration     int
	Lyric        string
	Note         midi.Note
	Intensity    float32
	PreUtterance *float32
	VoiceOverlap *float32
	StartPoint   *float32
	Envelope     Curve
	PitchBend    Curve
}

type CurveInterpolation uint8

const (
	CurveInterpolationLinear CurveInterpolation = iota
	CurveInterpolationSine
	CurveInterpolationRigid
	CurveInterpolationJump
)

type Curve []CurvePoint

type CurvePoint struct {
	umath.XY[float32]
	Interp CurveInterpolation
}

type Sequencer interface {
	Sequence() Sequence
}
