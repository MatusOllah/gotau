package umath

import (
	"cmp"
	"math"
	"slices"

	"github.com/SladkyCitron/gotau/umath/internal/ease"
)

// CurveInterpolation is the type of interpolation used between curve points.
type CurveInterpolation uint8

const (
	CurveInterpolationLinear CurveInterpolation = iota
	CurveInterpolationSineIn
	CurveInterpolationSineOut
	CurveInterpolationSineInOut
	CurveInterpolationSineOutIn
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

// At calculates and returns the value at the position in milliseconds.
func (c Curve) At(x float64) float64 {
	if len(c) == 0 || x < c[0].X || x > c[len(c)-1].X {
		return math.NaN()
	}

	// find start and end points that x sits in between
	i, ok := slices.BinarySearchFunc(c, CurvePoint{X: x}, func(a, b CurvePoint) int {
		return cmp.Compare(a.X, b.X)
	})
	if ok {
		return c[i].Y
	}
	if i == 0 || i >= len(c) {
		return math.NaN()
	}
	start := c[i-1]
	end := c[i]

	dx := end.X - start.X
	if dx <= 0 {
		return start.Y
	}
	t := x - start.X

	switch start.Interp {
	case CurveInterpolationLinear:
		return ease.Linear(t, start.Y, end.Y-start.Y, dx)
	case CurveInterpolationSineIn:
		return ease.SineIn(t, start.Y, end.Y-start.Y, dx)
	case CurveInterpolationSineOut:
		return ease.SineOut(t, start.Y, end.Y-start.Y, dx)
	case CurveInterpolationSineInOut:
		return ease.SineInOut(t, start.Y, end.Y-start.Y, dx)
	case CurveInterpolationSineOutIn:
		return ease.SineOutIn(t, start.Y, end.Y-start.Y, dx)
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
