package umath_test

import (
	"math"
	"testing"

	"github.com/SladkyCitron/gotau/umath"
	"github.com/stretchr/testify/assert"
)

func TestCurve_At(t *testing.T) {
	c := umath.Curve{
		{X: 0, Y: 0, Interp: umath.CurveInterpolationLinear},
		{X: 10, Y: 10},
	}

	assert.True(t, math.IsNaN(umath.Curve{}.At(0)))
	assert.True(t, math.IsNaN(c.At(-1)))
	assert.True(t, math.IsNaN(c.At(c[len(c)-1].X+1)))

	tests := []struct {
		pos  float64
		want float64
	}{
		{pos: 0, want: 0},
		{pos: 5, want: 5},
		{pos: 10, want: 10},
	}

	for _, test := range tests {
		assert.InDelta(t, test.want, c.At(test.pos), 1e-9, "tick: %f", test.pos)
	}
}
