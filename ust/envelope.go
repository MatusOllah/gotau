package ust

import (
	"fmt"

	"github.com/MatusOllah/slicestrconv"
)

// Envelope represents a volume envelope.
type Envelope struct {
	P1    float64
	P2    float64
	P3    float64
	V1    float64
	V2    float64
	V3    float64
	V4    float64
	Extra []float64
}

// ParseEnvelope parses a string representing an [Envelope] in an UST note.
// The string is expected to contain at least 7 comma-separated floating-point numbers,
// with the first 7 mapping to P1–P3 and V1–V4, and the rest going into Extra.
//
// Example: "5,35,0,100,100,0,0" or "5,35,0,100,100,0,0,10,20"
func ParseEnvelope(s string) (*Envelope, error) {
	slicestrconv.OpeningBracket = ""
	slicestrconv.ClosingBracket = ""
	values, err := slicestrconv.ParseFloat64Slice(s, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to parse envelope string: %w", err)
	}

	if len(values) < 7 {
		return nil, fmt.Errorf("envelope string must contain at least 7 values, got %d", len(values))
	}

	env := &Envelope{
		P1: values[0],
		P2: values[1],
		P3: values[2],
		V1: values[3],
		V2: values[4],
		V3: values[5],
		V4: values[6],
	}

	if len(values) > 7 {
		env.Extra = values[7:]
	}

	return env, nil
}
