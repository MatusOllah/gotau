package ust_test

import (
	"fmt"
	"testing"

	"github.com/MatusOllah/gotau/umath"
	"github.com/MatusOllah/gotau/ust"
	"github.com/stretchr/testify/assert"
)

func TestPitchBendMode_String(t *testing.T) {
	assert.Equal(t, "PitchBendModeLinear", ust.PitchBendModeLinear.String())
	assert.Equal(t, "PitchBendModeSine", ust.PitchBendModeSine.String())
	assert.Equal(t, "PitchBendModeRigid", ust.PitchBendModeRigid.String())
	assert.Equal(t, "PitchBendModeJump", ust.PitchBendModeJump.String())
}

func TestPitchBendMode_RawString(t *testing.T) {
	assert.Equal(t, "l", ust.PitchBendModeLinear.RawString())
	assert.Equal(t, "s", ust.PitchBendModeSine.RawString())
	assert.Equal(t, "r", ust.PitchBendModeRigid.RawString())
	assert.Equal(t, "j", ust.PitchBendModeJump.RawString())
}

func TestParsePitchBendMode(t *testing.T) {
	tests := []struct {
		name          string
		s             string
		expectedMode  ust.PitchBendMode
		expectedError error
	}{
		{name: "PitchBendModeLinear", s: "l", expectedMode: ust.PitchBendModeLinear, expectedError: nil},
		{name: "PitchBendModeSine", s: "s", expectedMode: ust.PitchBendModeSine, expectedError: nil},
		{name: "PitchBendModeRigid", s: "r", expectedMode: ust.PitchBendModeRigid, expectedError: nil},
		{name: "PitchBendModeJump", s: "j", expectedMode: ust.PitchBendModeJump, expectedError: nil},
		{name: "Invalid", s: "x", expectedMode: 0, expectedError: fmt.Errorf("invalid pitch bend mode string: x")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mode, err := ust.ParsePitchBendMode(test.s)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedMode, mode)
		})
	}
}

func TestParsePitchBend(t *testing.T) {
	pb, err := ust.ParsePitchBend(
		"5",       // type
		"10;2",    // start
		"",        // pbs (ignored because start is set)
		"30,40",   // pbw
		"0.5,1.0", // pby
		"l,s",     // pbm
	)
	assert.NoError(t, err)
	assert.NotNil(t, pb)

	assert.Equal(t, 5, pb.Type)
	assert.Equal(t, 10.0, pb.Start.X)
	assert.Equal(t, 2.0, pb.Start.Y)
	assert.Equal(t, []float64{30, 40}, pb.Widths)
	assert.Equal(t, []float64{0.5, 1.0}, pb.Ys)
	assert.Equal(t, []ust.PitchBendMode{ust.PitchBendModeLinear, ust.PitchBendModeSine}, pb.Modes)
}

func TestParsePitchBend_InvalidType(t *testing.T) {
	_, err := ust.ParsePitchBend("not-a-number", "", "", "", "", "")
	assert.Error(t, err)
}

func TestParsePitchBend_EmptyFields(t *testing.T) {
	pb, err := ust.ParsePitchBend("", "", "0", "10", "0", "")
	assert.NoError(t, err)
	assert.Equal(t, 5, pb.Type)
	assert.Equal(t, 0.0, pb.Start.X)
	assert.Equal(t, 0.0, pb.Start.Y)
}

func TestPitchBend_Curve(t *testing.T) {
	tests := []struct {
		name     string
		pb       *ust.PitchBend
		expected int
		check    func(t *testing.T, curve []umath.XY[float64])
	}{
		{
			name: "LinearSegment",
			pb: &ust.PitchBend{
				Start:  umath.XY[float64]{X: 0, Y: 0},
				Widths: []float64{10},
				Ys:     []float64{10},
				Modes:  []ust.PitchBendMode{ust.PitchBendModeLinear},
			},
			expected: 11,
			check: func(t *testing.T, curve []umath.XY[float64]) {
				assert.Equal(t, umath.XY[float64]{X: 0, Y: 0}, curve[0])
				assert.Equal(t, umath.XY[float64]{X: 10, Y: 10}, curve[len(curve)-1])
			},
		},
		{
			name: "RigidSegment",
			pb: &ust.PitchBend{
				Start:  umath.XY[float64]{X: 5, Y: 3},
				Widths: []float64{10},
				Ys:     []float64{7},
				Modes:  []ust.PitchBendMode{ust.PitchBendModeRigid},
			},
			expected: 11,
			check: func(t *testing.T, curve []umath.XY[float64]) {
				for _, pt := range curve {
					assert.Equal(t, 3.0, pt.Y)
				}
			},
		},
		{
			name: "JumpSegment",
			pb: &ust.PitchBend{
				Start:  umath.XY[float64]{X: 0, Y: 0},
				Widths: []float64{10},
				Ys:     []float64{5},
				Modes:  []ust.PitchBendMode{ust.PitchBendModeJump},
			},
			expected: 11,
			check: func(t *testing.T, curve []umath.XY[float64]) {
				assert.Equal(t, 0.0, curve[0].Y)
				assert.Equal(t, 5.0, curve[1].Y)
			},
		},
		{
			name: "SineSegment",
			pb: &ust.PitchBend{
				Start:  umath.XY[float64]{X: 0, Y: 0},
				Widths: []float64{10},
				Ys:     []float64{10},
				Modes:  []ust.PitchBendMode{ust.PitchBendModeSine},
			},
			expected: 11,
			check: func(t *testing.T, curve []umath.XY[float64]) {
				assert.Greater(t, curve[5].Y, 4.0)
			},
		},
		{
			name: "DefaultSegment",
			pb: &ust.PitchBend{
				Start:  umath.XY[float64]{X: 0, Y: 0},
				Widths: []float64{10},
				Ys:     []float64{10},
				Modes:  []ust.PitchBendMode{},
			},
			expected: 11,
			check: func(t *testing.T, curve []umath.XY[float64]) {
				assert.Equal(t, umath.XY[float64]{X: 0, Y: 0}, curve[0])
				assert.Equal(t, umath.XY[float64]{X: 10, Y: 10}, curve[len(curve)-1])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			curve := test.pb.Curve()
			assert.Len(t, curve, test.expected)
			test.check(t, curve)
		})
	}
}

func BenchmarkPitchBend_Curve(b *testing.B) {
	pb := &ust.PitchBend{
		Start:  umath.XY[float64]{X: 0, Y: 0},
		Widths: make([]float64, 100),
		Ys:     make([]float64, 100),
		Modes:  make([]ust.PitchBendMode, 100),
	}

	for i := range 100 {
		pb.Widths[i] = 10
		pb.Ys[i] = float64(i)
		pb.Modes[i] = ust.PitchBendModeLinear
	}

	for b.Loop() {
		_ = pb.Curve()
	}
}
