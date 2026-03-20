package gotau

import (
	"io"
	"log"
	"math"

	"github.com/SladkyCitron/gotau/sequence"
	"github.com/SladkyCitron/gotau/voicebank"
	"github.com/SladkyCitron/resona/freq"
)

// Synth is the main voice synthsizer that renders notes into audio samples.
type Synth struct {
	sched *scheduler
	vb    *voicebank.Voicebank
	sr    int
	buf   []float32
}

func New(sr freq.Frequency, vb *voicebank.Voicebank) *Synth {
	s := &Synth{
		sched: &scheduler{},
		vb:    vb,
		sr:    int(sr.Hertz()),
		buf:   make([]float32, 0, 8192),
	}
	return s
}

// SetResolution sets the timing resolution in ticks per quarter note (TPQN).
//
// Higher values increase timing precision but may result in more scheduling
// overhead.
func (s *Synth) SetResolution(resolution int) {
	s.sched.tpqn = resolution
}

// SetTempo sets the playback tempo in beats per minute (BPM).
func (s *Synth) SetTempo(tempo float32) {
	s.sched.bpm = tempo
}

// Enqueue adds notes to the synthesis queue.
//
// Notes are scheduled according to their tick position and will be rendered
// in order during subsequent ReadSamples calls.
func (s *Synth) Enqueue(notes ...sequence.Note) {
	s.sched.enqueue(notes...)
}

// EnqueueSequence adds all notes from the given sequence to the synthesis
// queue and updates the synthesizer's timing parameters.
//
// The sequence's resolution and tempo override the current scheduler settings.
func (s *Synth) EnqueueSequence(seq sequence.Sequence) {
	s.SetResolution(seq.Metadata.Resolution)
	s.SetTempo(seq.Metadata.Tempo)
	s.Enqueue(seq.Notes...)
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
