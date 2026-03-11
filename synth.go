package gotau

import (
	"io"
	"log"
	"math"

	"github.com/SladkyCitron/gotau/sequence"
	"github.com/SladkyCitron/gotau/voicebank"
)

type Synth struct {
	sched *scheduler
	vb    *voicebank.Voicebank
	sr    int
	buf   []float32
}

func New(vb *voicebank.Voicebank, seq sequence.Sequence) *Synth {
	s := &Synth{
		sched: &scheduler{tpqn: seq.Metadata.Resolution, bpm: seq.Metadata.Tempo},
		vb:    vb,
		sr:    44100, // sample rate temporary hardcoded
		buf:   make([]float32, 0, 8192),
	}
	s.Enqueue(seq.Notes...)
	return s
}

func (s *Synth) Enqueue(notes ...sequence.Note) {
	s.sched.enqueue(notes...)
}

func (s *Synth) ReadSamples(p []float32) (int, error) {
	n := 0

	// drain the buffer
	for n < len(p) && len(s.buf) > 0 {
		copied := copy(p[n:], s.buf)
		s.buf = s.buf[copied:]
		n += copied
	}

	// fill the buffer
	for n < len(p) {
		if len(s.sched.queue) == 0 {
			if n == 0 {
				return 0, io.EOF
			}
			return n, nil
		}

		notes := s.sched.pop(float64(len(p)-n) / float64(s.sr))
		for _, note := range notes {
			// emit silence before note
			if note.Position > s.sched.tickPos {
				s.debugLog("silence", note)
				buf := make([]float32, int(s.sched.ticksToSeconds(note.Position-s.sched.tickPos)*float64(s.sr)))
				s.buf = append(s.buf, buf...)
				s.sched.tickPos = note.Position
			}

			// render note
			s.debugLog("note", note)
			buf := make([]float32, int(s.sched.ticksToSeconds(note.Duration)*float64(s.sr)))
			freq := 440.0 * math.Pow(2, (float64(note.Note)-69)/12)
			step := 2 * math.Pi * freq / float64(s.sr)
			phase := 0.0
			for i := range buf {
				buf[i] = float32(math.Sin(phase))
				phase += step
			}
			s.buf = append(s.buf, buf...)
			s.sched.tickPos += note.Duration
		}

		copied := copy(p[n:], s.buf)
		s.buf = s.buf[copied:]
		n += copied
	}
	return n, nil
}

func (s *Synth) debugLog(msg string, note sequence.Note) {
	log.Printf("at %v -> %s: %v", s.sched.tickPos, msg, note)
}
