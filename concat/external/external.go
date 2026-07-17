// Package external implements an external concatenator.
package external

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/SladkyCitron/gotau/concat"
	"github.com/SladkyCitron/gotau/internal/timeutil"
	"github.com/SladkyCitron/resona/afmt"
	"github.com/SladkyCitron/resona/aio"
	"github.com/SladkyCitron/resona/codec/wav"
	"github.com/zeebo/xxh3"
)

var _ concat.Concatenator = (*Concatenator)(nil)

// Concatenator is a concatenator that uses an external command-line
// UTAU wavtool program to perform concatenation.
type Concatenator struct {
	// ConfigureCmd is an optional hook that allows configuring the
	// exec.Cmd before running it.
	ConfigureCmd func(cmd *exec.Cmd)

	cmdName   string
	sampleFmt afmt.SampleFormat
}

// New creates a new [Concatenator] with the given program name and
// sample format for encoding temporary WAV files for passing into the concatenator.
//
// The program should be a command-line UTAU wavtool (e.g. wavtool, wavtool-yawu, etc.)
// that accepts tail and note WAV file paths and
// other parameters as arguments and concatenates accordingly.
func New(name string, sampleFmt afmt.SampleFormat) *Concatenator {
	return &Concatenator{cmdName: name, sampleFmt: sampleFmt}
}

func (c *Concatenator) Concatenate(tail []float32, note []float32, cfg concat.ConcatenateConfig) ([]float32, error) {
	if len(cfg.Envelope) < 4 || len(cfg.Envelope) > 5 {
		return nil, errors.New("envelope curve must have 4 or 5 points")
	}

	tailWav, noteWav, err := c.createTempWav(tail, note, cfg)
	if err != nil {
		return nil, fmt.Errorf("external: failed to create temporary wav files: %w", err)
	}

	var length string
	// most wavtools have TPQN 480 hardcoded so if our TPQN is 480 we use the wavtool length syntax
	// and if it's anything other than 480 then we calculate the length here instead
	if cfg.Resolution == 480 {
		// ticks@bpm+-correction
		length = fmt.Sprintf("%d@%.0f%+.0f", cfg.Duration, cfg.Tempo, cfg.LengthDelta)
	} else {
		length = strconv.FormatFloat((timeutil.TicksToSeconds(cfg.Duration, cfg.Resolution, cfg.Tempo)*1000)+cfg.LengthDelta, 'f', -1, 64)
	}

	args := []string{
		tailWav,                                  // output file
		noteWav,                                  // input file
		strconv.FormatInt(int64(cfg.Offset), 10), // STP
		length,                                   // note length
		strconv.FormatInt(int64(cfg.Envelope[0].X), 10),     // p1
		strconv.FormatInt(int64(cfg.Envelope[1].X), 10),     // p2
		strconv.FormatInt(int64(cfg.Envelope[2].X), 10),     // p3
		strconv.FormatInt(int64(cfg.Envelope[0].Y*100), 10), // v1
		strconv.FormatInt(int64(cfg.Envelope[1].Y*100), 10), // v2
		strconv.FormatInt(int64(cfg.Envelope[2].Y*100), 10), // v3
		strconv.FormatInt(int64(cfg.Envelope[3].Y*100), 10), // v4
		strconv.FormatInt(int64(cfg.Overlap), 10),           // overlap
		strconv.FormatInt(int64(cfg.Envelope[3].X), 10),     // p4
	}
	if len(cfg.Envelope) >= 5 {
		args = append(
			args,
			strconv.FormatInt(int64(cfg.Envelope[4].X), 10),     // p5
			strconv.FormatInt(int64(cfg.Envelope[4].Y*100), 10), // v5
		)
	}
	cmd := exec.Command(c.cmdName, args...)
	if c.ConfigureCmd != nil {
		c.ConfigureCmd(cmd)
	}
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("external: failed to run wavtool command: %w", err)
	}

	if err := os.Remove(noteWav); err != nil {
		return nil, fmt.Errorf("external: failed to remove temporary note wav file: %w", err)
	}

	if _, err := os.Stat(tailWav); err == nil || !errors.Is(err, os.ErrNotExist) {
		// wav path
		f, err := os.Open(tailWav)
		if err != nil {
			return nil, fmt.Errorf("external: failed to open temporary output wav: %w", err)
		}
		deco, err := wav.NewDecoder(f)
		if err != nil {
			return nil, fmt.Errorf("external: failed to decode temporary output wav: %w", err)
		}
		out, err := aio.ReadAll(deco)
		if err != nil {
			return nil, fmt.Errorf("external: failed to read temporary output wav: %w", err)
		}
		if err := f.Close(); err != nil {
			return nil, fmt.Errorf("external: failed to close temporary output wav: %w", err)
		}
		if err := os.Remove(f.Name()); err != nil {
			return nil, fmt.Errorf("external: failed to remove temporary output wav: %w", err)
		}
		return out, nil
	} else {
		// whd + dat path
		whdPath := tailWav[:len(tailWav)-len(filepath.Ext(tailWav))] + ".whd"
		datPath := tailWav[:len(tailWav)-len(filepath.Ext(tailWav))] + ".dat"

		whdFile, err := os.Open(whdPath)
		if err != nil {
			return nil, fmt.Errorf("external: failed to open temporary whd: %w", err)
		}
		datFile, err := os.Open(datPath)
		if err != nil {
			return nil, fmt.Errorf("external: failed to open temporary dat: %w", err)
		}
		r := io.MultiReader(whdFile, datFile)
		deco, err := wav.NewDecoder(r)
		if err != nil {
			return nil, fmt.Errorf("external: failed to decode output wav: %w", err)
		}
		out, err := aio.ReadAll(deco)
		if err != nil {
			return nil, fmt.Errorf("external: failed to read output wav: %w", err)
		}
		if err := whdFile.Close(); err != nil {
			return nil, fmt.Errorf("external: failed to close temporary whd: %w", err)
		}
		if err := datFile.Close(); err != nil {
			return nil, fmt.Errorf("external: failed to close temporary dat: %w", err)
		}
		if err := os.Remove(whdFile.Name()); err != nil {
			return nil, fmt.Errorf("external: failed to remove temporary whd: %w", err)
		}
		if err := os.Remove(datFile.Name()); err != nil {
			return nil, fmt.Errorf("external: failed to remove temporary dat: %w", err)
		}
		return out, nil
	}
}

func (c *Concatenator) createTempWav(tail []float32, note []float32, cfg concat.ConcatenateConfig) (string, string, error) {
	// create filename
	h := xxh3.New()
	_, _ = h.WriteString(c.cmdName)
	_ = binary.Write(h, binary.LittleEndian, tail)
	_ = binary.Write(h, binary.LittleEndian, note)
	_ = binary.Write(h, binary.LittleEndian, cfg.Offset)
	_ = binary.Write(h, binary.LittleEndian, int64(cfg.Duration))
	_ = binary.Write(h, binary.LittleEndian, cfg.Tempo)
	_ = binary.Write(h, binary.LittleEndian, int64(cfg.Resolution))
	_ = binary.Write(h, binary.LittleEndian, cfg.LengthDelta)
	_ = binary.Write(h, binary.LittleEndian, cfg.Overlap)
	_ = binary.Write(h, binary.LittleEndian, uint64(len(cfg.Envelope)))
	for i := range cfg.Envelope {
		_ = binary.Write(h, binary.LittleEndian, cfg.Envelope[i].X)
		_ = binary.Write(h, binary.LittleEndian, cfg.Envelope[i].Y)
	}
	sum := h.Sum64()
	tailPath := filepath.Join(os.TempDir(), fmt.Sprintf("gotau-externalconcat-tail-%016x.wav", sum))
	notePath := filepath.Join(os.TempDir(), fmt.Sprintf("gotau-externalconcat-note-%016x.wav", sum))

	var wavFormat uint16
	switch c.sampleFmt.Encoding {
	case afmt.SampleEncodingInt, afmt.SampleEncodingUint:
		wavFormat = wav.FormatInt
	case afmt.SampleEncodingFloat:
		wavFormat = wav.FormatFloat
	default:
		return "", "", fmt.Errorf("invalid sample format: %s", c.sampleFmt.String())
	}

	tailFile, err := os.Create(tailPath)
	if err != nil {
		return "", "", err
	}
	defer tailFile.Close()
	tailEnc, err := wav.NewEncoder(tailFile, cfg.AudioFormat, c.sampleFmt, wavFormat)
	if err != nil {
		return "", "", err
	}
	if _, err := tailEnc.WriteSamples(tail); err != nil {
		return "", "", err
	}
	if err := tailEnc.Close(); err != nil {
		return "", "", err
	}

	noteFile, err := os.Create(notePath)
	if err != nil {
		return "", "", err
	}
	defer noteFile.Close()
	noteEnc, err := wav.NewEncoder(noteFile, cfg.AudioFormat, c.sampleFmt, wavFormat)
	if err != nil {
		return "", "", err
	}
	if _, err := noteEnc.WriteSamples(note); err != nil {
		return "", "", err
	}
	if err := noteEnc.Close(); err != nil {
		return "", "", err
	}

	return tailPath, notePath, nil
}
