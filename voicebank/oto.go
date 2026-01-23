package voicebank

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

// OtoEntry represents a single entry in an oto.ini file.
type OtoEntry struct {
	// Filename is the name of the audio file associated with this Oto entry.
	Filename string

	// Alias is the phoneme or alias associated with this Oto entry.
	Alias string

	// Offset is the offset time in milliseconds.
	Offset float32

	// Consonant is the consonant duration in milliseconds.
	Consonant float32

	// Cutoff is the cutoff time in milliseconds.
	Cutoff float32

	// Preutterance is the preutterance time in milliseconds.
	Preutterance float32

	// Overlap is the overlap time in milliseconds.
	Overlap float32
}

// Why does oto.ini and Oto (the audio thingie) have to have the same name...?! 😭
// It makes things so confusing...
//
// ... anyway...

// Oto represents the oto.ini configuration in an UTAU voicebank.
// It holds a list of phonemes, aliases, and their associated parameters.
type Oto []OtoEntry

type otoConfig struct {
	encoding       encoding.Encoding
	comment        rune
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

// OtoWithComment specifies the comment character to use when reading the oto.ini file.
// Lines beginning with this character without preceding whitespace will be ignored.
func OtoWithComment(comment rune) OtoOption {
	return func(cfg *otoConfig) {
		cfg.comment = comment
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
		comment:  '#',
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// we can't use ini package here because oto.ini contains
	// duplicate keys (i.e. filenames) which ini would reject

	scan := bufio.NewScanner(transform.NewReader(r, cfg.encoding.NewDecoder()))

	var oto Oto
	for scan.Scan() {
		line := scan.Text()
		if cfg.comment != 0 && strings.HasPrefix(strings.TrimSpace(line), string(cfg.comment)) {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		filename := parts[0]
		values := strings.Split(parts[1], ",")

		// filename=alias,offset,consonant,cutoff,preutter,overlap
		// filename and alias are strings, the rest are floats
		if len(values) != 6 {
			return oto, fmt.Errorf("invalid oto entry for %s, expected 6 values, got %d", strconv.Quote(filename), len(values))
		}

		alias := values[0]

		if values[1] == "" {
			values[1] = "0"
		}
		offset, err := strconv.ParseFloat(values[1], 32)
		if err != nil {
			return oto, fmt.Errorf("failed to parse offset value for %s: %w", strconv.Quote(filename), err)
		}

		if values[2] == "" {
			values[2] = "0"
		}
		consonant, err := strconv.ParseFloat(values[2], 32)
		if err != nil {
			return oto, fmt.Errorf("failed to parse consonant value for %s: %w", strconv.Quote(filename), err)
		}

		if values[3] == "" {
			values[3] = "0"
		}
		cutoff, err := strconv.ParseFloat(values[3], 32)
		if err != nil {
			return oto, fmt.Errorf("failed to parse cutoff value for %s: %w", strconv.Quote(filename), err)
		}

		if values[4] == "" {
			values[4] = "0"
		}
		preutter, err := strconv.ParseFloat(values[4], 32)
		if err != nil {
			return oto, fmt.Errorf("failed to parse preutterance value for %s: %w", strconv.Quote(filename), err)
		}

		if values[5] == "" {
			values[5] = "0"
		}
		overlap, err := strconv.ParseFloat(values[5], 32)
		if err != nil {
			return oto, fmt.Errorf("failed to parse overlap value for %s: %w", strconv.Quote(filename), err)
		}

		oto = append(oto, OtoEntry{
			Filename:     filename,
			Alias:        alias,
			Offset:       float32(offset),
			Consonant:    float32(consonant),
			Cutoff:       float32(cutoff),
			Preutterance: float32(preutter),
			Overlap:      float32(overlap),
		})
	}
	return oto, scan.Err()
}

//TODO: optimize with binary search and lookup table; also gotta make Oto a struct with a slice inside instead of just a slice

// Get retrieves an [OtoEntry] by its phoneme alias.
func (o Oto) Get(alias string) (_ OtoEntry, ok bool) {
	for i := range o {
		if o[i].Alias == alias {
			return o[i], true
		}
	}
	return OtoEntry{}, false
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
		fmtStr := "%s=%s,%W.Pf,%W.Pf,%W.Pf,%W.Pf,%W.Pf\n"
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
