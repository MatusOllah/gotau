package ust

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Envelope represents a volume envelope.
type Envelope struct {
	P1         float64 // P1 is the fade-in start offset (in milliseconds).
	P2         float64 // P2 is the fade-in end offset (in milliseconds).
	P3         float64 // P3 is the decay start offset (in milliseconds).
	V1         float64 // V1 is the volume at P1 (in %).
	V2         float64 // V2 is the volume at P2 (in %).
	V3         float64 // V3 is the volume at P3 (in %).
	V4         float64 // V4 is the volume at P4 (in %).
	P4         float64 // P4 is the sustain end offset (in milliseconds).
	P5         float64 // P5 is the fade-out end offset (in milliseconds).
	V5         float64 // V5 is the volume at P5 (in %).
	HasRelease bool    // HasRelease indicates whether the envelope has a release (i.e. P4 and P5 are defined).
}

// ParseEnvelope parses a string representing an [Envelope] in an UST note.
func ParseEnvelope(s string) (*Envelope, error) {
	parts := strings.Split(s, ",")

	if len(parts) < 7 {
		return nil, fmt.Errorf("envelope string must contain at least 7 values, got %d", len(parts))
	}
	if len(parts) > 12 {
		return nil, fmt.Errorf("envelope string must contain at most 11 values, got %d", len(parts))
	}

	vals := make([]float64, len(parts))
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if p == "%" {
			vals[i] = math.NaN() // use NaN to indicate separator
			continue
		}
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			//return nil, fmt.Errorf("invalid envelope value at %d: %w", i, err)
			vals[i] = 0
			continue
		}
		vals[i] = v
	}

	env := &Envelope{
		P1: vals[0],
		P2: vals[1],
		P3: vals[2],
		V1: vals[3],
		V2: vals[4],
		V3: vals[5],
		V4: vals[6],
	}

	if len(parts) >= 11 && strings.TrimSpace(parts[7]) == "%" {
		env.HasRelease = true
		env.P4 = vals[8]
		env.P5 = vals[9]
		env.V5 = vals[10]
	}

	return env, nil
}
