package pitch_test

import (
	"math/rand/v2"
	"testing"

	"github.com/SladkyCitron/gotau/pitch"
	"github.com/stretchr/testify/assert"
)

func TestEncodeResamplerPitchBendString(t *testing.T) {
	tests := []struct {
		name string
		x    []float64
		want string
	}{
		{
			name: "empty",
			x:    []float64{},
			want: "AA",
		},
		{
			name: "single point",
			x:    []float64{1},
			want: "AB",
		},
		{
			name: "run length",
			x:    []float64{0, 0, 0, 0},
			want: "AA#3#",
		},
		{
			name: "run length with 2 values",
			x:    []float64{0, 0, 0, 0, 1, 1, 1, 1},
			want: "AA#3#AB#3#",
		},
		{
			name: "max",
			x:    []float64{2047},
			want: "f/",
		},
		{
			name: "min",
			x:    []float64{-2048},
			want: "gA",
		},
		{
			name: "clamping",
			x:    []float64{-3000, 3000},
			want: "gAf/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, pitch.EncodeResamplerPitchBendString(tt.x))
		})
	}
}

func BenchmarkEncodeResamplerPitchBendString(b *testing.B) {
	data := make([]float64, 1000)
	for i := range data {
		data[i] = rand.Float64()*4096 - 2048
	}
	b.ResetTimer()
	for b.Loop() {
		_ = pitch.EncodeResamplerPitchBendString(data)
	}
}
