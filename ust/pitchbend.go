package ust

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/MatusOllah/gotau/umath"
	"github.com/MatusOllah/slicestrconv"
)

// PitchBendMode represents a pitch bend segment interpolation mode.
type PitchBendMode int

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
	Start  umath.XY[float64] // Start is the starting point in ticks (X-axis) and initial pitch offset (Y-axis in semitones).
	Widths []float64         // Widths are the widths in ticks for each pitch segment.
	Ys     []float64         // Ys are the pitch offsets in semitones for each segment.
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
	pb.Widths, err = slicestrconv.ParseFloat64Slice(pbw, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pitch bend widths: %w", err)
	}

	// PBY
	pb.Ys, err = slicestrconv.ParseFloat64Slice(pby, 10)
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
	x, err := strconv.ParseFloat(pbsParts[0], 64)
	if err != nil {
		return err
	}
	y := 0.0
	if len(pbsParts) > 1 {
		y, err = strconv.ParseFloat(pbsParts[1], 64)
		if err != nil {
			return err
		}
	}
	pb.Start = umath.XY[float64]{X: x, Y: y}
	return nil
}

// Curve computes a curve from the pitch bend data.
func (pb *PitchBend) Curve() []umath.XY[float64] {
	const samplesPerSegment = 10
	curve := []umath.XY[float64]{}
	prevX, prevY := pb.Start.X, pb.Start.Y

	// Handle cases where we have incomplete data
	segmentCount := iMin(len(pb.Widths), len(pb.Ys))
	for i := range segmentCount {
		width := pb.Widths[i]
		endY := pb.Ys[i]
		endX := prevX + width

		mode := PitchBendModeLinear
		if i < len(pb.Modes) {
			mode = pb.Modes[i]
		}

		for j := 0; j <= samplesPerSegment; j++ {
			t := float64(j) / float64(samplesPerSegment)
			var y float64

			switch mode {
			case PitchBendModeLinear:
				y = prevY*(1-t) + endY*t
			case PitchBendModeSine:
				y = prevY + (endY-prevY)*(0.5-0.5*math.Cos(math.Pi*t))
			case PitchBendModeRigid:
				y = prevY
			case PitchBendModeJump:
				if j == 0 {
					y = prevY
				} else {
					y = endY
				}
			default:
				y = prevY // fallback to flat segment
			}

			x := prevX + width*t
			curve = append(curve, umath.XY[float64]{X: x, Y: y})
		}

		prevX = endX
		prevY = endY
	}

	return curve
}

func iMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
