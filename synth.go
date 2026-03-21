package gotau

import (
	"fmt"
	"io"
	"log"

	"github.com/SladkyCitron/gotau/phonemizer"
	"github.com/SladkyCitron/gotau/sequence"
	"github.com/SladkyCitron/gotau/voicebank"
	"github.com/SladkyCitron/resona/codec"
	_ "github.com/SladkyCitron/resona/codec/au"
	_ "github.com/SladkyCitron/resona/codec/qoa"
	_ "github.com/SladkyCitron/resona/codec/wav"
)

const startBufSize = 4096 // Size of initial allocation for buffer

// Synth is the main voice synthsizer that renders notes into audio samples.
type Synth struct {
	vb        *voicebank.Voicebank
	ph        phonemizer.Phonemizer
	sched     *scheduler
	sr        int
	buf       []float32
	prevLyric string
}

func New(sr int, vb *voicebank.Voicebank) *Synth {
	s := &Synth{
		vb:    vb,
		ph:    &phonemizer.Default{},
		sched: &scheduler{},
		sr:    sr,
		buf:   make([]float32, 0, startBufSize),
	}
	return s
}

// Buffer controls memory allocation by the Synth.
// It sets the internal buffer to use when rendering notes.
// The contents of the buffer are ignored.
func (s *Synth) Buffer(buf []float32) {
	s.buf = buf[0:cap(buf)]
}

// SetPhonemizer sets the phonemizer.
func (s *Synth) SetPhonemizer(ph phonemizer.Phonemizer) {
	s.ph = ph
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

		for note := range s.sched.popSeq(float64(len(p)-n) / float64(s.sr)) {
			if err := s.renderNote(note); err != nil {
				copied := copy(p[n:], s.buf)
				s.buf = s.buf[copied:]
				n += copied
				return n, fmt.Errorf("gotau Synth: failed to render note %q: %w", note.Lyric, err)
			}
		}

		copied := copy(p[n:], s.buf)
		s.buf = s.buf[copied:]
		n += copied
	}
	return n, nil
}

func (s *Synth) renderNote(note sequence.Note) error {
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

	otoEntry, ok := s.getOtoEntry(note)
	if !ok {
		// oto entry not found; emit silence instead
		s.buf = append(s.buf, buf...)
		s.sched.tickPos += note.Duration
		s.prevLyric = note.Lyric
		return nil
	}

	f, err := s.vb.FS().Open(otoEntry.FilePath())
	if err != nil {
		return err
	}

	deco, _, err := codec.Decode(f)
	if err != nil {
		return err
	}
	if sr := int(deco.Format().SampleRate.Hertz()); sr != s.sr {
		return fmt.Errorf("voicebank (%d Hz) and synth (%d Hz) sample rate do not match", sr, s.sr)
	}

	if _, err := deco.ReadSamples(buf); err != nil {
		return err
	}

	s.buf = append(s.buf, buf...)
	s.sched.tickPos += note.Duration
	s.prevLyric = note.Lyric
	return nil
}

func (s *Synth) getOtoEntry(note sequence.Note) (e voicebank.OtoEntry, ok bool) {
	resolveCfg := phonemizer.ResolveConfig{
		PrevLyric: s.prevLyric,
		Lyric:     note.Lyric,
		Note:      note.Note,
	}
	for alias := range s.ph.Resolve(resolveCfg) {
		e, ok = s.vb.Oto.Get(alias)
		if ok {
			return e, true
		}
	}
	return voicebank.OtoEntry{}, false
}

func (s *Synth) debugLog(msg string, note sequence.Note) {
	log.Printf("at %v -> %s: %v", s.sched.tickPos, msg, note)
}
