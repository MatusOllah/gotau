package concat

import (
	"errors"

	"github.com/SladkyCitron/resona/dsp"
)

// Default is the default [Concatenator]. It mimics original UTAU wavtool behavior.
type Default struct {
	env []float32
}

/*
func iclamp(value, min, max int) int {
	switch {
	case value < min:
		return min
	case value > max:
		return max
	default:
		return value
	}
}
*/

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

	offsetSamples := int(cfg.Offset / 1000 * float64(cfg.AudioFormat.SampleRate.Hertz()))
	lengthSamples := int(cfg.Length / 1000 * float64(cfg.AudioFormat.SampleRate.Hertz()))
	overlapSamples := int(cfg.Overlap / 1000 * float64(cfg.AudioFormat.SampleRate.Hertz()))

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
	if cap(d.env) < len(buf) {
		d.env = make([]float32, len(buf))
	} else {
		d.env = d.env[:len(buf)]
	}
	d.interpolateEnvelope(cfg.Envelope[0].X, cfg.Envelope[0].Y, cfg.Envelope[1].X, cfg.Envelope[1].Y)
	d.interpolateEnvelope(cfg.Envelope[1].X, cfg.Envelope[1].Y, cfg.Envelope[2].X, cfg.Envelope[2].Y)
	d.interpolateEnvelope(cfg.Envelope[2].X, cfg.Envelope[2].Y, cfg.Envelope[3].X, cfg.Envelope[3].Y)
	d.interpolateEnvelope(cfg.Envelope[3].X, cfg.Envelope[3].Y, cfg.Envelope[4].X, cfg.Envelope[4].Y)

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
