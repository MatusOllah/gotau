package voicebank

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

//TODO: make these float32s

// OtoEntry represents a single entry in an oto.ini file.
type OtoEntry struct {
	Filename     string
	Alias        string
	Offset       float64
	Consonant    float64
	Cutoff       float64
	Preutterance float64
	Overlap      float64
}

// Why does oto.ini and Oto (the audio thingie) have to have the same name...?! 😭
// It makes things so confusing...
//
// ... anyway...

// Oto (not to be confused with the [audio playback library]) represents
// the oto.ini configuration in an UTAU voicebank.
// It holds a list of phonemes, aliases, and their associated parameters.
//
// [audio playback library]: https://github.com/ebitengine/oto
type Oto []OtoEntry

type otoConfig struct {
	encoding       encoding.Encoding
	floatWidth     int
	floatPercision int
}

// OtoOption represents an option for passing into oto.ini-related functions and methods.
type OtoOption func(*otoConfig)

// OtoWithEncoding specifies the character encoding to use when reading or writing the oto.ini file.
func OtoWithEncoding(encoding encoding.Encoding) OtoOption {
	return func(cfg *otoConfig) {
		cfg.encoding = encoding
	}
}

// OtoWithFloatFormat specifies the float formatting options (width and precision)
// to use when encoding float values in the oto.ini file.
func OtoWithFloatFormat(width, precision int) OtoOption {
	if width < 0 {
		panic("float width cannot be negative")
	}

	if precision < 0 {
		panic("float precision cannot be negative")
	}

	return func(cfg *otoConfig) {
		cfg.floatWidth = width
		cfg.floatPercision = precision
	}
}

// DecodeOto parses and decodes an oto.ini file from the provided [io.Reader].
func DecodeOto(r io.Reader, opts ...OtoOption) (Oto, error) {
	cfg := &otoConfig{
		encoding: encoding.Nop,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	scan := bufio.NewScanner(transform.NewReader(r, cfg.encoding.NewDecoder()))

	var oto Oto
	for scan.Scan() {
		line := scan.Text()
		parts := strings.SplitN(line, "=", 2)
		filename := parts[0]
		values := strings.Split(parts[1], ",")

		// filename=alias,offset,consonant,cutoff,preutter,overlap
		// filename and alias are strings, the rest are floats
		if len(values) != 6 {
			return oto, fmt.Errorf("invalid oto entry for %s, expected 6 values, got %d", strconv.Quote(filename), len(values))
		}

		alias := values[0]

		offset, err := strconv.ParseFloat(values[1], 64)
		if err != nil {
			return oto, fmt.Errorf("failed to parse offset value for %s: %w", strconv.Quote(filename), err)
		}

		consonant, err := strconv.ParseFloat(values[2], 64)
		if err != nil {
			return oto, fmt.Errorf("failed to parse consonant value for %s: %w", strconv.Quote(filename), err)
		}

		cutoff, err := strconv.ParseFloat(values[3], 64)
		if err != nil {
			return oto, fmt.Errorf("failed to parse cutoff value for %s: %w", strconv.Quote(filename), err)
		}

		preutter, err := strconv.ParseFloat(values[4], 64)
		if err != nil {
			return oto, fmt.Errorf("failed to parse preutterance value for %s: %w", strconv.Quote(filename), err)
		}

		overlap, err := strconv.ParseFloat(values[5], 64)
		if err != nil {
			return oto, fmt.Errorf("failed to parse overlap value for %s: %w", strconv.Quote(filename), err)
		}

		oto = append(oto, OtoEntry{
			Filename:     filename,
			Alias:        alias,
			Offset:       offset,
			Consonant:    consonant,
			Cutoff:       cutoff,
			Preutterance: preutter,
			Overlap:      overlap,
		})
	}
	return oto, nil
}

// Get retrieves an [OtoEntry] by its phoneme alias.
func (o Oto) Get(alias string) (_ OtoEntry, ok bool) {
	idx, ok := slices.BinarySearchFunc(o, OtoEntry{Alias: alias}, func(a, b OtoEntry) int {
		return strings.Compare(a.Alias, b.Alias)
	})
	if !ok {
		return OtoEntry{}, false
	}
	return o[idx], true
}

// Encode encodes and writes the oto.ini entries to the provided [io.Writer].
func (o Oto) Encode(w io.Writer, opts ...OtoOption) error {
	cfg := &otoConfig{
		encoding: encoding.Nop,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	newWriter := transform.NewWriter(w, cfg.encoding.NewEncoder())

	for _, entry := range o {
		// filename=alias,offset,consonant,cutoff,preutter,overlap
		// filename and alias are strings, the rest are floats
		fmtStr := "%s=%s,%W.Pf,%W.Pf,%W.Pf,%W.Pf,%W.Pf"
		fmtStr = strings.ReplaceAll(fmtStr, "W", strconv.Itoa(cfg.floatWidth))
		fmtStr = strings.ReplaceAll(fmtStr, "P", strconv.Itoa(cfg.floatPercision))
		_, err := fmt.Fprintf(newWriter, fmtStr,
			entry.Filename,
			entry.Alias,
			entry.Offset,
			entry.Consonant,
			entry.Cutoff,
			entry.Preutterance,
			entry.Overlap,
		)
		if err != nil {
			return fmt.Errorf("failed to write oto entry for %s: %w", strconv.Quote(entry.Filename), err)
		}
	}
	return nil
}
