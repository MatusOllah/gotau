package sequence

import (
	"cmp"
	"math"
	"slices"
)

// CurveInterpolation is the type of interpolation used between curve points.
type CurveInterpolation uint8

const (
	CurveInterpolationLinear CurveInterpolation = iota
	CurveInterpolationSine
	CurveInterpolationRigid
	CurveInterpolationJump
)

// Curve represents a curve. It consists of a list of curve points and the interpolation type between them.
// The curve points must be sorted by [CurvePoint.X] in ascending order and must not have duplicate X values.
type Curve []CurvePoint

// CurvePoint represents a single point on a curve.
type CurvePoint struct {
	// X is the position in MIDI ticks.
	X int

	// Y is the value.
	Y float64

	// Interp is the interpolation type to the next point. If it's the last point, it is ignored.
	Interp CurveInterpolation
}

var cmpFn = func(a, b CurvePoint) int { return cmp.Compare(a.X, b.X) }

// At calculates and returns the value at the tick.
func (c Curve) At(tick int) float64 {
	if len(c) == 0 || tick < c[0].X || tick > c[len(c)-1].X {
		return math.NaN()
	}

	// exact match shortcut
	i, ok := slices.BinarySearchFunc(c, CurvePoint{X: tick}, cmpFn)
	if ok {
		return c[i].Y
	}

	// find start and end points that tick sits in between
	i, _ = slices.BinarySearchFunc(c, CurvePoint{X: tick}, cmpFn)
	if i == 0 || i >= len(c) {
		return math.NaN()
	}
	start := c[i-1]
	end := c[i]

	dx := end.X - start.X
	if dx <= 0 {
		return start.Y
	}
	t := float64(tick-start.X) / float64(dx)

	switch start.Interp {
	case CurveInterpolationLinear:
		return lerp(start.Y, end.Y, t)
	case CurveInterpolationSine:
		return lerp(start.Y, end.Y, (1-math.Cos(math.Pi*t))/2)
	case CurveInterpolationRigid:
		if tick < end.X {
			return start.Y
		}
		return end.Y
	case CurveInterpolationJump:
		return end.Y
	default:
		return math.NaN()
	}
}

func lerp(a, b, t float64) float64 {
	return a + t*(b-a)
}
