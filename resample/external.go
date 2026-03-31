package resample

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/SladkyCitron/gotau/pitch"
	"github.com/SladkyCitron/resona/afmt"
	"github.com/SladkyCitron/resona/aio"
	"github.com/SladkyCitron/resona/codec/wav"
	"github.com/zeebo/xxh3"
)

var _ Resampler = (*ExternalResampler)(nil)

// ExternalResampler is a resampler that uses an external command-line UTAU resampler program to perform resampling.
type ExternalResampler struct {
	cmd       *exec.Cmd
	sampleFmt afmt.SampleFormat
}

// NewExternal creates a new [ExternalResampler] with the given program name and
// sample format for encoding temporary WAV files for passing into the resampler.
//
// The program should be a command-line UTAU resampler (e.g. resampler, moresampler, straycat, etc.)
// that accepts input and output WAV file paths and other parameters as arguments and processes the input WAV file accordingly.
func NewExternal(name string, sampleFmt afmt.SampleFormat) *ExternalResampler {
	return &ExternalResampler{cmd: exec.Command(name), sampleFmt: sampleFmt}
}

// Cmd returns the command that will be executed for resampling.
func (r *ExternalResampler) Cmd() *exec.Cmd {
	return r.cmd
}

func (r *ExternalResampler) Resample(in aio.SampleReader, cfg ResampleConfig) (aio.SampleReader, error) {
	input, err := r.createTempWav(in, cfg)
	if err != nil {
		return nil, fmt.Errorf("resample ExternalResampler: failed to create temporary wav file: %w", err)
	}

	output := input[:len(input)-len(filepath.Ext(input))] + "-out.wav"

	r.cmd.Args = append([]string(nil),
		input,
		output,
		strconv.FormatInt(int64(cfg.Pitch), 10),
		strconv.FormatFloat(cfg.Velocity, 'f', -1, 64),
		cfg.Flags,
		strconv.FormatFloat(cfg.Offset, 'f', -1, 64),
		strconv.FormatFloat(cfg.Length, 'f', -1, 64),
		strconv.FormatFloat(cfg.Consonant, 'f', -1, 64),
		strconv.FormatFloat(cfg.Cutoff, 'f', -1, 64),
		strconv.FormatFloat(cfg.Volume, 'f', -1, 64),
		strconv.FormatFloat(cfg.Modulation, 'f', -1, 64),
		strconv.FormatFloat(cfg.Tempo, 'f', -1, 64),
		pitch.EncodeResamplerPitchBendString(cfg.PitchBend, cfg.Pitch, cfg.Length, cfg.Tempo, cfg.Resolution),
	)
	if err := r.cmd.Run(); err != nil {
		return nil, fmt.Errorf("resample ExternalResampler: failed to run resampler command: %w", err)
	}

	if err := os.Remove(input); err != nil {
		return nil, fmt.Errorf("resample ExternalResampler: failed to remove temporary wav file: %w", err)
	}

	out, err := r.decodeOutFile(output)
	if err != nil {
		return nil, fmt.Errorf("resample ExternalResampler: failed to decode output wav file: %w", err)
	}

	return out, nil
}

func (r *ExternalResampler) createTempWav(in aio.SampleReader, cfg ResampleConfig) (string, error) {
	// create filename
	h := xxh3.New()
	h.WriteString(r.cmd.Path)
	h.Write([]byte{byte(cfg.Pitch)})
	binary.Write(h, binary.LittleEndian, cfg.Velocity)
	h.WriteString(cfg.Flags)
	binary.Write(h, binary.LittleEndian, cfg.Offset)
	binary.Write(h, binary.LittleEndian, cfg.Length)
	binary.Write(h, binary.LittleEndian, cfg.Consonant)
	binary.Write(h, binary.LittleEndian, cfg.Cutoff)
	binary.Write(h, binary.LittleEndian, cfg.Volume)
	binary.Write(h, binary.LittleEndian, cfg.Modulation)
	binary.Write(h, binary.LittleEndian, cfg.Tempo)
	binary.Write(h, binary.LittleEndian, uint64(cfg.Resolution))
	binary.Write(h, binary.LittleEndian, uint64(len(cfg.PitchBend)))
	for _, pt := range cfg.PitchBend {
		binary.Write(h, binary.LittleEndian, uint64(pt.X))
		binary.Write(h, binary.LittleEndian, pt.Y)
		h.Write([]byte{byte(pt.Interp)})
	}
	path := filepath.Join(os.TempDir(), fmt.Sprintf("gotau-externalresampler-%016x.wav", h.Sum64()))

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var wavFormat uint16
	switch r.sampleFmt.Encoding {
	case afmt.SampleEncodingInt, afmt.SampleEncodingUint:
		wavFormat = wav.FormatInt
	case afmt.SampleEncodingFloat:
		wavFormat = wav.FormatFloat
	default:
		return "", fmt.Errorf("invalid sample format: %s", r.sampleFmt.String())
	}
	enc, err := wav.NewEncoder(f, cfg.AudioFormat, r.sampleFmt, wavFormat)
	if err != nil {
		return "", err
	}

	if _, err := aio.Copy(enc, in); err != nil {
		return "", err
	}

	if err := enc.Close(); err != nil {
		return "", err
	}

	f.Sync()

	return path, nil
}

func (r *ExternalResampler) decodeOutFile(path string) (aio.SampleReader, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	deco, err := wav.NewDecoder(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	if err := os.Remove(path); err != nil {
		return nil, err
	}

	return deco, nil
}
