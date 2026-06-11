package concat

import (
	"errors"
	"math"

	"github.com/SladkyCitron/resona/dsp"
	"github.com/SladkyCitron/resona/freq"
)

var _ Concatenator = (*Windowed)(nil)

// Windowed is the default [Concatenator].
type Windowed struct {
	env []float32
}

func msToSamples(ms float64, sr freq.Frequency) int {
	return int(math.Round(ms / 1000 * sr.Hertz()))
}

func (c *Windowed) interpolateEnvelope(startX int, startY float64, endX int, endY float64) {
	if startX == endX || startX >= len(c.env) {
		return
	}
	if startX < 0 {
		startX = 0
	}
	if endX > len(c.env) {
		endX = len(c.env)
	}
	dx := float64(endX - startX)
	dy := endY - startY
	for i := startX; i < endX; i++ {
		c.env[i] = float32(startY + dy*float64(i-startX)/dx)
	}
}

func (c *Windowed) Concatenate(tail []float32, note []float32, cfg ConcatenateConfig) ([]float32, error) {
	if len(cfg.Envelope) < 4 || len(cfg.Envelope) > 5 {
		return nil, errors.New("envelope curve must have 4 or 5 points")
	}

	offsetSamples := msToSamples(cfg.Offset, cfg.AudioFormat.SampleRate)
	lengthSamples := msToSamples(cfg.Length, cfg.AudioFormat.SampleRate)
	overlapSamples := min(msToSamples(cfg.Overlap, cfg.AudioFormat.SampleRate), len(tail))
	if offsetSamples < 0 {
		offsetSamples = 0
	}

	if offsetSamples >= len(note) {
		// return silence if offset is past length
		return make([]float32, lengthSamples), nil
	}

	trueNote := note[offsetSamples:]
	buf := make([]float32, lengthSamples)

	if len(c.env) < lengthSamples {
		c.env = make([]float32, lengthSamples)
	} else {
		c.env = c.env[:lengthSamples]
	}

	// interpolate envelope
	p1 := msToSamples(cfg.Envelope[0].X, cfg.AudioFormat.SampleRate)
	p2 := msToSamples(cfg.Envelope[1].X, cfg.AudioFormat.SampleRate)
	p3 := msToSamples(cfg.Envelope[2].X, cfg.AudioFormat.SampleRate)
	p4 := msToSamples(cfg.Envelope[3].X, cfg.AudioFormat.SampleRate)
	var p5 int
	if len(cfg.Envelope) == 5 {
		p5 = msToSamples(cfg.Envelope[4].X, cfg.AudioFormat.SampleRate)
	}
	c.interpolateEnvelope(0, 0, p1, cfg.Envelope[0].Y)
	c.interpolateEnvelope(p1, cfg.Envelope[0].Y, p1+p2, cfg.Envelope[1].Y)
	if len(cfg.Envelope) == 5 {
		c.interpolateEnvelope(p1+p2, cfg.Envelope[1].Y, p1+p2+p5, cfg.Envelope[4].Y)
		c.interpolateEnvelope(p1+p2+p5, cfg.Envelope[4].Y, lengthSamples-p4-p3, cfg.Envelope[2].Y)
	} else {
		c.interpolateEnvelope(p1+p2, cfg.Envelope[1].Y, p2, cfg.Envelope[2].Y)
		c.interpolateEnvelope(p2, cfg.Envelope[2].Y, lengthSamples-p4-p3, cfg.Envelope[2].Y)
		//d.interpolateEnvelope(p1+p2, cfg.Envelope[1].Y, lengthSamples-p4-p3, cfg.Envelope[2].Y)
	}
	c.interpolateEnvelope(lengthSamples-p4-p3, cfg.Envelope[2].Y, lengthSamples-p4, cfg.Envelope[3].Y)
	c.interpolateEnvelope(lengthSamples-p4, cfg.Envelope[3].Y, lengthSamples, 0)

	// apply envelope and crossfade
	w := hann(overlapSamples * 2)
	for i := range buf {
		var sample float32
		if i < len(trueNote) {
			sample = trueNote[i] * c.env[i]
		}
		if i < len(tail) && i < overlapSamples {
			tailSample := tail[len(tail)-overlapSamples+i]
			sample = tailSample*(1-w[i]) + sample*float32(w[i])
		}
		buf[i] = dsp.Clamp(sample)
	}

	return buf, nil
}

// https://github.com/SladkyCitron/resona/blob/main/dsp/window/window.go#L104
// but float32

// hann returns an n-point Hann window.
//
// For n == 1, the window is defined as [1.0].
//
// Reference: https://en.wikipedia.org/wiki/Window_function#Hann_window
func hann(n int) []float32 {
	if n <= 0 {
		return nil
	}

	w := make([]float32, n)

	// Special case
	if n == 1 {
		w[0] = 1
		return w
	}

	for i := range w {
		w[i] = 0.5 * (1 - float32(math.Cos((2*math.Pi*float64(i))/float64(n-1))))
	}
	return w
}
