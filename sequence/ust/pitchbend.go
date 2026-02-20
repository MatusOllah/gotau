package ust

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/SladkyCitron/gotau/umath"
	"github.com/SladkyCitron/slicestrconv"
)

// PitchBendMode represents a pitch bend segment interpolation mode.
type PitchBendMode uint8

const (
	PitchBendModeLinear PitchBendMode = iota
	PitchBendModeSine
	PitchBendModeRigid
	PitchBendModeJump
)

// String returns a string representation of the pitch bend mode.
func (m PitchBendMode) String() string {
	switch m {
	case PitchBendModeLinear:
		return "PitchBendModeLinear"
	case PitchBendModeSine:
		return "PitchBendModeSine"
	case PitchBendModeRigid:
		return "PitchBendModeRigid"
	case PitchBendModeJump:
		return "PitchBendModeJump"
	default:
		panic("invalid pitch bend mode")
	}
}

// RawString returns a raw UST string representation of the pitch bend mode.
func (m PitchBendMode) RawString() string {
	switch m {
	case PitchBendModeLinear:
		return "l"
	case PitchBendModeSine:
		return "s"
	case PitchBendModeRigid:
		return "r"
	case PitchBendModeJump:
		return "j"
	default:
		panic("invalid pitch bend mode")
	}
}

// ParsePitchBendMode parses a pitch bend mode string.
func ParsePitchBendMode(s string) (PitchBendMode, error) {
	switch s {
	case "l":
		return PitchBendModeLinear, nil
	case "s":
		return PitchBendModeSine, nil
	case "r":
		return PitchBendModeRigid, nil
	case "j":
		return PitchBendModeJump, nil
	default:
		return 0, fmt.Errorf("invalid pitch bend mode string: %s", s)
	}
}

// PitchBend represents the pitch bend data.
type PitchBend struct {
	Type   int               // Type is the pitch bend type (0 = no bend, 5 = default).
	Start  umath.XY[float32] // Start is the starting point in ticks (X-axis) and initial pitch offset (Y-axis in semitones).
	Widths []float32         // Widths are the widths in ticks for each pitch segment.
	Ys     []float32         // Ys are the pitch offsets in semitones for each segment.
	Modes  []PitchBendMode   // Modes are the interpolation modes for each segment.
}

func ParsePitchBend(typ, start, pbs, pbw, pby, pbm string) (pb *PitchBend, err error) {
	pb = &PitchBend{}

	// PBType
	pb.Type = 5
	if typ != "" {
		pb.Type, err = strconv.Atoi(typ)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pitch bend type: %w", err)
		}
	}

	// PBStart / PBS
	if start != "" {
		if err := pb.parseStart(start); err != nil {
			return nil, fmt.Errorf("failed to parse pitch bend start: %w", err)
		}
	} else if pbs != "" {
		if err := pb.parseStart(pbs); err != nil {
			return nil, fmt.Errorf("failed to parse pitch bend start: %w", err)
		}
	}

	slicestrconv.OpeningBracket = ""
	slicestrconv.ClosingBracket = ""

	// PBW
	pb.Widths, err = slicestrconv.ParseFloat32Slice(pbw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pitch bend widths: %w", err)
	}

	// PBY
	pb.Ys, err = slicestrconv.ParseFloat32Slice(pby)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pitch bend ys: %w", err)
	}

	// PBM
	for m := range strings.SplitSeq(pbm, ",") {
		if m == "" {
			continue
		}
		mode, err := ParsePitchBendMode(m)
		if err != nil {
			return nil, err
		}
		pb.Modes = append(pb.Modes, mode)
	}

	return pb, nil
}

func (pb *PitchBend) parseStart(s string) error {
	pbsParts := strings.Split(s, ";")
	x, err := strconv.ParseFloat(pbsParts[0], 32)
	if err != nil {
		return err
	}
	y := 0.0
	if len(pbsParts) > 1 {
		y, err = strconv.ParseFloat(pbsParts[1], 32)
		if err != nil {
			return err
		}
	}
	pb.Start = umath.XY[float32]{X: float32(x), Y: float32(y)}
	return nil
}
