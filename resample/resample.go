package resample

import (
	"github.com/SladkyCitron/gotau/sequence"
	"github.com/SladkyCitron/gotau/voicebank"
	"github.com/SladkyCitron/resona/aio"
)

type Resampler interface {
	Resample(in aio.SampleReader, cfg ResampleConfig) (aio.SampleReader, error)
}

type ResampleConfig struct {
	Note       sequence.Note
	OtoEntry   voicebank.OtoEntry
	Length     float64
	Volume     float64
	Modulation float64
	PitchBend  string
}
