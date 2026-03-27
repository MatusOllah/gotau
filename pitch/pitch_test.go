package pitch_test

import (
	"testing"

	"github.com/SladkyCitron/gotau/pitch"
	"github.com/SladkyCitron/gotau/sequence"
	"github.com/stretchr/testify/assert"
)

func TestEncodeResamplerPitchBend(t *testing.T) {
	curve := sequence.Curve{
		{X: 0, Y: 0, Interp: sequence.CurveInterpolationLinear},
		{X: 960, Y: 0},
	}

	got := pitch.EncodeResamplerPitchBend(curve, 0, 1, 120, 480)

	assert.Equal(t, got, []byte("AA#200#"))
}

func BenchmarkEncodeResamplerPitchBend(b *testing.B) {
	curve := sequence.Curve{
		{X: 0, Y: 10, Interp: sequence.CurveInterpolationLinear},
		{X: 960, Y: 20},
	}
	for b.Loop() {
		_ = pitch.EncodeResamplerPitchBend(curve, 60, 1, 120, 480)
	}
}
