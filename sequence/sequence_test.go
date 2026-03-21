package sequence_test

import (
	"testing"
	"time"

	"github.com/SladkyCitron/gotau/sequence"
	"github.com/stretchr/testify/assert"
)

func TestSequence_Len(t *testing.T) {
	seq := sequence.Sequence{
		Metadata: sequence.Metadata{
			Resolution: 480,
			Tempo:      120,
		},
		Notes: []sequence.Note{
			{Position: 0, Duration: 480},
			{Position: 480, Duration: 480},
			{Position: 960, Duration: 960},
		},
	}
	assert.Equal(t, 1920, seq.Len())
}

func TestSequence_Duration(t *testing.T) {
	seq := sequence.Sequence{
		Metadata: sequence.Metadata{
			Resolution: 480,
			Tempo:      120,
		},
		Notes: []sequence.Note{
			{Position: 0, Duration: 480},
			{Position: 480, Duration: 480},
			{Position: 960, Duration: 960},
		},
	}
	assert.Equal(t, 2*time.Second, seq.Duration())
}
