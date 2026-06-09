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
	// X is the position in milliseconds.
	X float64

	// Y is the value.
	Y float64

	// Interp is the interpolation type to the next point. If it's the last point, it is ignored.
	Interp CurveInterpolation
}

var cmpFn = func(a, b CurvePoint) int { return cmp.Compare(a.X, b.X) }

// At calculates and returns the value at the position in milliseconds.
func (c Curve) At(x float64) float64 {
	if len(c) == 0 || x < c[0].X || x > c[len(c)-1].X {
		return math.NaN()
	}

	// exact match shortcut
	i, ok := slices.BinarySearchFunc(c, CurvePoint{X: x}, cmpFn)
	if ok {
		return c[i].Y
	}

	// find start and end points that tick sits in between
	i, _ = slices.BinarySearchFunc(c, CurvePoint{X: x}, cmpFn)
	if i == 0 || i >= len(c) {
		return math.NaN()
	}
	start := c[i-1]
	end := c[i]

	dx := end.X - start.X
	if dx <= 0 {
		return start.Y
	}
	t := (x - start.X) / dx

	switch start.Interp {
	case CurveInterpolationLinear:
		return lerp(start.Y, end.Y, t)
	case CurveInterpolationSine:
		return lerp(start.Y, end.Y, (1-math.Cos(math.Pi*t))/2)
	case CurveInterpolationRigid:
		if t < 0.5 {
			return start.Y
		}
		return end.Y
	case CurveInterpolationJump:
		return end.Y
	default:
		return math.NaN()
	}
}

// AtClamped is the same as [Curve.At], but it clamps the input to the curve's domain instead of returning NaN.
func (c Curve) AtClamped(x float64) float64 {
	if len(c) == 0 {
		return 0
	}
	if x <= c[0].X {
		return c[0].Y
	}
	if x >= c[len(c)-1].X {
		return c[len(c)-1].Y
	}
	return c.At(x)
}

func lerp(a, b, t float64) float64 {
	return a + t*(b-a)
}
