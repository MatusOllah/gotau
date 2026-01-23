package voicebank

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"gitlab.com/gomidi/midi/v2"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

// PrefixMap represents the prefix.map configuration in an UTAU voicebank.
// It maps notes to their corresponding [Prefix] (prefix and suffix).
type PrefixMap map[midi.Note]Prefix

// Prefix represents a single entry in the prefix.map file.
// It holds the prefix and suffix.
type Prefix struct {
	Prefix string
	Suffix string
}

type prefixMapConfig struct {
	encoding  encoding.Encoding
	delimiter rune
	comment   rune
}

// PrefixMapOption represents an option for passing into prefix.map-related functions and methods.
type PrefixMapOption func(*prefixMapConfig)

// PrefixMapWithEncoding specifies the character encoding to use when reading or writing the prefix.map file.
func PrefixMapWithEncoding(enc encoding.Encoding) PrefixMapOption {
	return func(cfg *prefixMapConfig) {
		cfg.encoding = enc
	}
}

// PrefixMapWithDelimiter specifies the delimiter to use when reading or writing the prefix.map file.
func PrefixMapWithDelimiter(delim rune) PrefixMapOption {
	return func(cfg *prefixMapConfig) {
		cfg.delimiter = delim
	}
}

// PrefixMapWithComment specifies the comment character to use when reading the prefix.map file.
// Lines beginning with this character without preceding whitespace will be ignored.
func PrefixMapWithComment(comment rune) PrefixMapOption {
	return func(cfg *prefixMapConfig) {
		cfg.comment = comment
	}
}

// DecodePrefixMap parses and decodes a prefix.map file from the provided reader.
func DecodePrefixMap(r io.Reader, opts ...PrefixMapOption) (PrefixMap, error) {
	cfg := &prefixMapConfig{
		encoding:  encoding.Nop,
		delimiter: '\t',
		comment:   '#',
	}

	for _, opt := range opts {
		opt(cfg)
	}

	scan := bufio.NewScanner(transform.NewReader(r, cfg.encoding.NewDecoder()))

	prefixMap := make(PrefixMap)
	for scan.Scan() {
		line := scan.Text()
		if cfg.comment != 0 && strings.HasPrefix(strings.TrimSpace(line), string(cfg.comment)) {
			continue
		}

		parts := strings.SplitN(line, string(cfg.delimiter), 3)
		if len(parts) < 3 {
			return nil, fmt.Errorf("voicebank prefix.map: invalid line: %q", line)
		}

		note, err := parseNote(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("voicebank prefix.map: invalid note %q: %w", parts[0], err)
		}

		prefixMap[note] = Prefix{
			Prefix: strings.TrimSpace(parts[1]),
			Suffix: strings.TrimSpace(parts[2]),
		}
	}
	return prefixMap, scan.Err()
}

// we use A-H to support German notation; H means B natural
// https://en.wikipedia.org/wiki/Musical_note#B%E2%99%AD,_B_and_H

var noteRe = regexp.MustCompile(`(?i)([A-H])([#b])?-?(\d+)`)

func parseNote(s string) (midi.Note, error) {
	m := noteRe.FindStringSubmatch(s)
	if m == nil {
		return 0, fmt.Errorf("invalid note format: %q", s)
	}

	base := strings.ToUpper(m[1])
	accidental := strings.ToLower(m[2])
	octave, err := strconv.Atoi(m[3])
	if err != nil {
		return 0, fmt.Errorf("invalid octave: %q", m[3])
	}

	switch base + accidental {
	case "C":
		return midi.Note(midi.C(uint8(octave))), nil
	case "C#", "Db":
		return midi.Note(midi.Db(uint8(octave))), nil
	case "D":
		return midi.Note(midi.D(uint8(octave))), nil
	case "D#", "Eb":
		return midi.Note(midi.Eb(uint8(octave))), nil
	case "E":
		return midi.Note(midi.E(uint8(octave))), nil
	case "F":
		return midi.Note(midi.F(uint8(octave))), nil
	case "F#", "Gb":
		return midi.Note(midi.Gb(uint8(octave))), nil
	case "G":
		return midi.Note(midi.G(uint8(octave))), nil
	case "G#", "Ab":
		return midi.Note(midi.Ab(uint8(octave))), nil
	case "A":
		return midi.Note(midi.A(uint8(octave))), nil
	case "A#", "Bb", "Hb": // H = B in German notation
		return midi.Note(midi.Bb(uint8(octave))), nil
	case "B", "H": // H = B in German notation
		return midi.Note(midi.B(uint8(octave))), nil
	default:
		return 0, fmt.Errorf("invalid base: %q", base+accidental)
	}
}
