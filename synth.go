package gotau

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"path/filepath"
	"strings"

	"github.com/SladkyCitron/gotau/cache"
	"github.com/SladkyCitron/gotau/concat"
	"github.com/SladkyCitron/gotau/phonemizer"
	"github.com/SladkyCitron/gotau/resample"
	"github.com/SladkyCitron/gotau/sequence"
	"github.com/SladkyCitron/gotau/voicebank"
	"github.com/SladkyCitron/resona/afmt"
	"github.com/SladkyCitron/resona/aio"
	"github.com/SladkyCitron/resona/codec"
	_ "github.com/SladkyCitron/resona/codec/au"
	_ "github.com/SladkyCitron/resona/codec/qoa"
	"github.com/SladkyCitron/resona/codec/wav"
	"github.com/SladkyCitron/resona/freq"
)

const startBufSize = 4096 // Size of initial allocation for buffer

// Synth is the main singing voice synthsizer that renders notes into audio samples.
type Synth struct {
	vb        *voicebank.Voicebank
	ph        phonemizer.Phonemizer
	res       resample.Resampler
	cat       concat.Concatenator
	resCache  cache.Cache
	sched     *scheduler
	sr        int
	buf       []float32
	vbFileBuf bytes.Buffer
	prevLyric string
	nextLyric string
}

// New creates a new [Synth] with the given sample rate, voicebank, resampler, and concatenator.
func New(sr int, vb *voicebank.Voicebank, res resample.Resampler, cat concat.Concatenator) *Synth {
	s := &Synth{
		vb:       vb,
		ph:       &phonemizer.Default{},
		res:      res,
		cat:      cat,
		resCache: &cache.NopCache{},
		sched:    &scheduler{},
		sr:       sr,
		buf:      make([]float32, 0, startBufSize),
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

// SetResampleCache sets the cache for storing resampled notes.
func (s *Synth) SetResampleCache(c cache.Cache) {
	s.resCache = c
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
		s.debugLog("fallback silence", note)
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
	defer f.Close()

	s.vbFileBuf.Reset()
	if _, err := io.Copy(&s.vbFileBuf, f); err != nil {
		return err
	}

	deco, _, err := codec.Decode(bytes.NewReader(s.vbFileBuf.Bytes()))
	if err != nil {
		return err
	}
	if sr := int(deco.Format().SampleRate.Hertz()); sr != s.sr {
		return fmt.Errorf("voicebank (%d Hz) and synth (%d Hz) sample rate do not match", sr, s.sr)
	}

	newLength = math.Ceil((newLength+s.getStartPoint(note)+25)/50) * 5000
	resampleCfg := resample.ResampleConfig{
		Pitch:       note.Note,
		Velocity:    s.getVelocity(note),
		Flags:       note.Flags,
		Offset:      otoEntry.Offset,
		Length:      newLength,
		Consonant:   otoEntry.Consonant,
		Cutoff:      otoEntry.Cutoff,
		Intensity:   note.Intensity,
		Modulation:  note.Modulation,
		Tempo:       s.sched.bpm,
		Resolution:  s.sched.tpqn,
		PitchBend:   note.PitchBend,
		AudioFormat: afmt.Format{SampleRate: freq.Frequency(s.sr) * freq.Hertz, NumChannels: 1},
	}

	var resampled aio.SampleReader
	key := s.getKeyFunc(resampleCfg)
	ctx := context.Background()
	if rc, err := s.resCache.Open(ctx, key); err == nil {
		resampled, err = wav.NewDecoder(rc)
		if err != nil {
			return err
		}
	} else {
		if analyzer, ok := s.res.(resample.Analyzer); ok {
			// check if there's the analysis sidecar file available
			ext := filepath.Ext(otoEntry.FilePath())
			name := otoEntry.FilePath()[:len(otoEntry.FilePath())-len(ext)]
			analysisPath := name + strings.ReplaceAll(ext, ".", "_") + analyzer.AnalysisExt()
			analysisFile, err := s.vb.FS().Open(analysisPath)
			if err == nil {
				resampled, err = analyzer.ResampleWithAnalysis(deco, analysisFile, resampleCfg)
				if err != nil {
					return fmt.Errorf("failed to resample: %w", err)
				}
				if err := analysisFile.Close(); err != nil {
					return fmt.Errorf("failed to close analysis sidecar file: %w", err)
				}
			} else {
				// nope
				resampled, err = s.res.Resample(deco, resampleCfg)
				if err != nil {
					return fmt.Errorf("failed to resample: %w", err)
				}
			}
		} else {
			resampled, err = s.res.Resample(deco, resampleCfg)
			if err != nil {
				return fmt.Errorf("failed to resample: %w", err)
			}
		}

		// cache the resampled audio
		f, err := s.resCache.Create(ctx, key)
		if err != nil {
			_ = f.Abort()
			return fmt.Errorf("failed to create cache entry: %w", err)
		}

		enc, err := wav.NewEncoder(
			f,
			resampleCfg.AudioFormat,
			afmt.SampleFormat{BitDepth: 32, Encoding: afmt.SampleEncodingFloat, Endian: binary.LittleEndian},
			wav.FormatFloat,
		)
		if err != nil {
			_ = f.Abort()
			return fmt.Errorf("failed to create wav encoder for caching: %w", err)
		}

		if _, err := aio.Copy(enc, resampled); err != nil {
			_ = f.Abort()
			return fmt.Errorf("failed to cache resampled audio: %w", err)
		}

		if err := enc.Close(); err != nil {
			_ = f.Abort()
			return fmt.Errorf("failed to close wav encoder for caching: %w", err)
		}

		if err := f.Close(); err != nil {
			_ = f.Abort()
			return fmt.Errorf("failed to close cache entry: %w", err)
		}
	}

	if _, err := resampled.ReadSamples(buf); err != nil && err != io.EOF {
		return fmt.Errorf("failed to read resampled audio: %w", err)
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

func (s *Synth) getPreutter(otoEntry voicebank.OtoEntry, note sequence.Note) float64 {
	if note.Preutterance != nil {
		return *note.Preutterance
	}
	return otoEntry.Preutterance
}

func (s *Synth) getVelocity(note sequence.Note) float64 {
	if note.Velocity != nil {
		return *note.Velocity
	}
	return 0
}

func (s *Synth) getStartPoint(note sequence.Note) float64 {
	if note.StartPoint != nil {
		return *note.StartPoint * math.Pow(2, 1-s.getVelocity(note)/100)
	}
	return 0
}

func (s *Synth) debugLog(msg string, note sequence.Note) {
	log.Printf("at %v -> %s: %v", s.sched.tickPos, msg, note)
}
