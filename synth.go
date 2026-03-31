package gotau

import (
	"fmt"
	"io"
	"log"

	"github.com/SladkyCitron/gotau/concat"
	"github.com/SladkyCitron/gotau/phonemizer"
	"github.com/SladkyCitron/gotau/resample"
	"github.com/SladkyCitron/gotau/sequence"
	"github.com/SladkyCitron/gotau/voicebank"
	"github.com/SladkyCitron/resona/codec"
	_ "github.com/SladkyCitron/resona/codec/au"
	_ "github.com/SladkyCitron/resona/codec/qoa"
	_ "github.com/SladkyCitron/resona/codec/wav"
)

const startBufSize = 4096 // Size of initial allocation for buffer

// Synth is the main singing voice synthsizer that renders notes into audio samples.
type Synth struct {
	vb        *voicebank.Voicebank
	ph        phonemizer.Phonemizer
	res       resample.Resampler
	cat       concat.Concatenator
	sched     *scheduler
	sr        int
	buf       []float32
	prevLyric string
	nextLyric string
}

// New creates a new [Synth] with the given sample rate, voicebank, resampler, and concatenator.
func New(sr int, vb *voicebank.Voicebank, res resample.Resampler, cat concat.Concatenator) *Synth {
	s := &Synth{
		vb:    vb,
		ph:    &phonemizer.Default{},
		res:   res,
		cat:   cat,
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
func (s *Synth) SetTempo(tempo float64) {
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

		seconds := float64(len(p)-n) / float64(s.sr)
		for note := range s.sched.popSeq(seconds) {
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
	// get next lyric
	if next, ok := s.sched.peek(); ok {
		s.nextLyric = next.Lyric
	} else {
		s.nextLyric = ""
	}

	// get oto
	otoEntry, otoOk := s.getOtoEntry(note)

	// get preutterance of current and next note
	preutter := s.getPreutter(otoEntry, note)
	preutterSec := preutter / 1000
	var nextPreutter float64
	if next, ok := s.sched.peek(); ok {
		if next.Preutterance != nil {
			nextPreutter = *next.Preutterance
		}
		if nextOtoEntry, ok := s.getOtoEntry(next); ok {
			nextPreutter = s.getPreutter(nextOtoEntry, next)
		}
	}
	nextPreutterSec := nextPreutter / 1000

	// emit possible silence before note
	if startTick := note.Position - s.sched.secondsToTicks(preutterSec); startTick > s.sched.tickPos {
		s.debugLog("silence", note)
		buf := make([]float32, int((s.sched.ticksToSeconds(note.Position-s.sched.tickPos)-preutterSec)*float64(s.sr)))
		s.buf = append(s.buf, buf...)
		s.sched.tickPos = startTick
	}

	s.debugLog("note", note)

	// oto entry not found; emit silence instead
	if !otoOk {
		buf := make([]float32, int((s.sched.ticksToSeconds(note.Duration)-nextPreutterSec)*float64(s.sr)))
		s.buf = append(s.buf, buf...)
		s.sched.tickPos += note.Duration
		s.prevLyric = note.Lyric
		return nil
	}

	// adjust actual note duration and timing
	var enroachment float64
	if _, ok := s.sched.peek(); ok && s.sched.ticksToSeconds(note.Duration) <= nextPreutterSec {
		enroachment = nextPreutterSec - s.sched.ticksToSeconds(note.Duration)
	}
	newLength := s.sched.ticksToSeconds(note.Duration) + preutterSec - enroachment

	buf := make([]float32, int(newLength*float64(s.sr)))

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

	/*
		resampleCfg := resample.ResampleConfig{
			Note:     note,
			OtoEntry: otoEntry,
		}
		s.res.Resample(deco, resampleCfg)
	*/

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

func (s *Synth) getPreutter(otoEntry voicebank.OtoEntry, note sequence.Note) float64 {
	if note.Preutterance != nil {
		return *note.Preutterance
	}
	return otoEntry.Preutterance
}

func (s *Synth) debugLog(msg string, note sequence.Note) {
	log.Printf("at %v -> %s: %v", s.sched.tickPos, msg, note)
}
