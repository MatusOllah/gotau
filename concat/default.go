package concat

import (
	"errors"

	"github.com/SladkyCitron/resona/dsp"
	"github.com/SladkyCitron/resona/freq"
)

// Default is the default [Concatenator]. It mimics original UTAU wavtool behavior.
type Default struct {
	env []float32
}

func msToSamples(ms float64, sr freq.Frequency) int {
	return int(ms / 1000 * sr.Hertz())
}

func (d *Default) interpolateEnvelope(startX int, startY float64, endX int, endY float64) {
	if startX == endX || startX >= len(d.env) {
		return
	}
	if startX < 0 {
		startX = 0
	}
	if endX > len(d.env) {
		endX = len(d.env)
	}
	dx := float64(endX - startX)
	dy := endY - startY
	for i := startX; i < endX; i++ {
		d.env[i] = float32(startY + dy*float64(i-startX)/dx)
	}
}

func (d *Default) Concatenate(tail []float32, note []float32, cfg ConcatenateConfig) ([]float32, error) {
	if len(cfg.Envelope) != 5 {
		return nil, errors.New("envelope curve must have 5 points")
	}

	offsetSamples := msToSamples(cfg.Offset, cfg.AudioFormat.SampleRate)
	lengthSamples := msToSamples(cfg.Length, cfg.AudioFormat.SampleRate)
	overlapSamples := msToSamples(cfg.Overlap, cfg.AudioFormat.SampleRate)
	if offsetSamples < 0 {
		offsetSamples = 0
	}
	if offsetSamples >= len(note) {
		// return silence if offset is past length
		return make([]float32, lengthSamples), nil
	}

	trueNote := note[offsetSamples:]
	buf := make([]float32, lengthSamples)

	if cap(d.env) < lengthSamples {
		d.env = make([]float32, lengthSamples)
	} else {
		d.env = d.env[:lengthSamples]
	}

	// interpolate envelope
	x0 := msToSamples(cfg.Envelope[0].X, cfg.AudioFormat.SampleRate)
	x1 := msToSamples(cfg.Envelope[1].X, cfg.AudioFormat.SampleRate)
	x2 := msToSamples(cfg.Envelope[2].X, cfg.AudioFormat.SampleRate)
	x3 := msToSamples(cfg.Envelope[3].X, cfg.AudioFormat.SampleRate)
	x4 := msToSamples(cfg.Envelope[4].X, cfg.AudioFormat.SampleRate)
	d.interpolateEnvelope(x0, cfg.Envelope[0].Y, x1, cfg.Envelope[1].Y)
	d.interpolateEnvelope(x1, cfg.Envelope[1].Y, x2, cfg.Envelope[2].Y)
	d.interpolateEnvelope(x2, cfg.Envelope[2].Y, x3, cfg.Envelope[3].Y)
	d.interpolateEnvelope(x3, cfg.Envelope[3].Y, x4, cfg.Envelope[4].Y)

	// apply envelope and crossfade
	for i := range buf {
		var sample float32
		if i < len(trueNote) {
			sample = trueNote[i]
		}
		sample *= d.env[i]
		if i < len(tail) && i < overlapSamples {
			sample += tail[i]
		}
		buf[i] = dsp.Clamp(sample)
	}

	return buf, nil
}
