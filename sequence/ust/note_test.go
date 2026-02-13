package ust_test

import (
	"testing"

	"github.com/SladkyCitron/gotau/sequence/ust"
	"github.com/stretchr/testify/assert"
)

func TestIsLyricRest(t *testing.T) {
	tests := []struct {
		name         string
		lyric        string
		expectedRest bool
	}{
		{name: "R", lyric: "R", expectedRest: true},
		{name: "Dash", lyric: "-", expectedRest: true},
		{name: "RDash", lyric: "R-", expectedRest: false},
		{name: "A", lyric: "a", expectedRest: false},
		{name: "ADash", lyric: "a-", expectedRest: false},
		{name: "AR", lyric: "aR", expectedRest: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isRest := ust.IsLyricRest(test.lyric)
			assert.Equal(t, test.expectedRest, isRest)
		})
	}
}
