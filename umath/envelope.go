package umath

// Envelope represents a volume envelope curve.
// X values map to P values and X values map to V values normalized to [0.0, 1.0].
//
// It should have 4 or 5 points.
// If it has 5 points, the last point represent the release phase (i.e. sustain end and fade-out).
type Envelope XYs[float64]
