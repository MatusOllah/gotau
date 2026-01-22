package ust_test

import (
	"testing"

	"github.com/MatusOllah/gotau/sequence/ust"
	"github.com/stretchr/testify/assert"
)

func TestParseEnvelope(t *testing.T) {
	tests := []struct {
		name        string
		s           string
		expectedEnv *ust.Envelope
		expectErr   bool
		errContains string
	}{
		{
			name:        "ValidBasic",
			s:           "5,35,0,100,100,0,0",
			expectedEnv: &ust.Envelope{P1: 5, P2: 35, P3: 0, V1: 100, V2: 100, V3: 0, V4: 0},
			expectErr:   false,
		},
		{
			name: "ValidWithExtra",
			s:    "1,2,3,4,5,6,7,8,9",
			expectedEnv: &ust.Envelope{
				P1: 1, P2: 2, P3: 3,
				V1: 4, V2: 5, V3: 6, V4: 7,
				Extra: []float64{8, 9},
			},
			expectErr: false,
		},
		{
			name:        "Invalid_NonNumericInput",
			s:           "1,2,3,wonderhoy,5,6,7",
			expectErr:   true,
			errContains: "failed to parse envelope string",
		},
		{
			name:        "Invalid_TooFewValues",
			s:           "1,2,3,4",
			expectErr:   true,
			errContains: "must contain at least 7 values",
		},
		{
			name:        "Invalid_EmptyString",
			s:           "",
			expectErr:   true,
			errContains: "failed to parse envelope string",
		},
		{
			name:        "Invalid_WeirdSpacing",
			s:           " 1 , 2 ,3, 4 ,5,6 , 7 ",
			expectedEnv: &ust.Envelope{P1: 1, P2: 2, P3: 3, V1: 4, V2: 5, V3: 6, V4: 7},
			expectErr:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, err := ust.ParseEnvelope(test.s)

			if test.expectErr {
				assert.Error(t, err)
				assert.Nil(t, env)
				if test.errContains != "" {
					assert.Contains(t, err.Error(), test.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedEnv, env)
			}
		})
	}
}
