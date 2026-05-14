package sequence_test

import (
	"math"
	"testing"

	"github.com/SladkyCitron/gotau/sequence"
	"github.com/stretchr/testify/assert"
)

func TestCurve_At(t *testing.T) {
	c := sequence.Curve{
		{X: 0, Y: 0, Interp: sequence.CurveInterpolationLinear},
		{X: 10, Y: 10, Interp: sequence.CurveInterpolationSine},
		{X: 20, Y: 5, Interp: sequence.CurveInterpolationRigid},
		{X: 30, Y: 20, Interp: sequence.CurveInterpolationJump},
		{X: 40, Y: 30},
	}

	assert.True(t, math.IsNaN(sequence.Curve{}.At(0)))
	assert.True(t, math.IsNaN(c.At(-1)))
	assert.True(t, math.IsNaN(c.At(c[len(c)-1].X+1)))

	tests := []struct {
		tick int
		want float64
	}{
		{tick: 0, want: 0},
		{tick: 5, want: 5},
		{tick: 10, want: 10},
		{tick: 15, want: 7.5},
		{tick: 20, want: 5},
		{tick: 25, want: 5},
		{tick: 30, want: 20},
		{tick: 35, want: 30},
		{tick: 40, want: 30},
	}

	for _, test := range tests {
		assert.InDelta(t, test.want, c.At(test.tick), 1e-9, "tick: %d", test.tick)
	}
}
