package ust

import (
	"fmt"
	"strconv"
	"strings"
)

// Envelope represents a volume envelope.
type Envelope struct {
	P1 EnvelopeValue
	P2 EnvelopeValue
	P3 EnvelopeValue
	V1 EnvelopeValue
	V2 EnvelopeValue
	V3 EnvelopeValue
	V4 EnvelopeValue
	P4 EnvelopeValue
	P5 EnvelopeValue
}

// EnvelopeValue represents a single value in an envelope, which can be either
// a fixed value or an automatic value (represented in UST as %).
type EnvelopeValue struct {
	Value float32
	Auto  bool
}

// Env is shorthand for &Envelope{...}.
func Env(p1, p2, p3, v1, v2, v3, v4, p4, p5 float32) *Envelope {
	return &Envelope{
		P1: EnvelopeValue{Value: p1},
		P2: EnvelopeValue{Value: p2},
		P3: EnvelopeValue{Value: p3},
		V1: EnvelopeValue{Value: v1},
		V2: EnvelopeValue{Value: v2},
		V3: EnvelopeValue{Value: v3},
		V4: EnvelopeValue{Value: v4},
		P4: EnvelopeValue{Value: p4},
		P5: EnvelopeValue{Value: p5},
	}
}

func parseEnvelopeValue(s string) (EnvelopeValue, error) {
	if s == "%" || s == "" {
		return EnvelopeValue{Auto: true}, nil
	}
	v, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return EnvelopeValue{}, err
	}
	return EnvelopeValue{Value: float32(v)}, nil
}

// ParseEnvelope parses a string representing an [Envelope] in an UST note.
// The string is expected to contain at least 7 comma-separated floating-point numbers.
func ParseEnvelope(s string) (*Envelope, error) {
	parts := strings.Split(s, ",")

	if len(parts) < 7 {
		return nil, fmt.Errorf("envelope string must contain at least 7 values, got %d", len(parts))
	}
	if len(parts) > 9 {
		return nil, fmt.Errorf("envelope string must contain at most 9 values, got %d", len(parts))
	}

	vals := make([]EnvelopeValue, len(parts))
	for i, p := range parts {
		ev, err := parseEnvelopeValue(strings.TrimSpace(p))
		if err != nil {
			return nil, fmt.Errorf("invalid envelope value at %d: %w", i, err)
		}
		vals[i] = ev
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

	if len(vals) == 8 {
		env.P4 = vals[7]
	}
	if len(vals) == 9 {
		env.P4 = vals[7]
		env.P5 = vals[8]
	}

	return env, nil
}
