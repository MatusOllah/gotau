package ust

import "gonum.org/v1/plot/plotter"

type PitchBendMode int

const (
	PitchBendModeLinear PitchBendMode = iota
	PitchBendModeSine
	PitchBendModeRigid
	PitchBendModeJump
)

// PitchBend represents the pitch bend data.
type PitchBend struct {
	Type   int             // Type is the pitch bend type (0 = no bend, 5 = default).
	Start  plotter.XY      // Start is the starting point in ticks (X-axis) and initial pitch offset (Y-axis in semitones).
	Widths []float64       // Widths are the widths in ticks for each pitch segment.
	Ys     []float64       // Ys are the pitch offsets in semitones for each segment.
	Modes  []PitchBendMode // Modes are the interpolation modes for each segment.
}

func ParsePitchBend(typ, start, pbs, pbw, pby, pbm string) (*PitchBend, error) {
	//TODO: parse pitch bend

	return nil, nil
}
