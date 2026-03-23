package voicebank

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

// OtoEntry represents a single entry in an oto.ini file.
type OtoEntry struct {
	// Filename is the name of the audio file associated with this Oto entry.
	Filename string

	// Directory is the path to the directory where the oto.ini file and audio files are located.
	Directory string

	// Alias is the phoneme or alias associated with this Oto entry.
	Alias string

	// Offset is the offset time in milliseconds.
	Offset float64

	// Consonant is the consonant duration in milliseconds.
	Consonant float64

	// Cutoff is the cutoff time in milliseconds.
	Cutoff float64

	// Preutterance is the preutterance time in milliseconds.
	Preutterance float64

	// Overlap is the overlap time in milliseconds.
	Overlap float64
}

// FilePath returns the file path to the audio file.
func (e OtoEntry) FilePath() string {
	// using path and not filepath because fs.FS uses forward slashes
	return path.Join(e.Directory, e.Filename)
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
	floatPrecision int
	dir            string
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

// OtoWithFloatPrecision specifies the float precision to use when writing float values in the oto.ini file.
func OtoWithFloatPrecision(prec int) OtoOption {
	if prec < 0 {
		panic("float precision cannot be negative")
	}

	return func(cfg *otoConfig) {
		cfg.floatPrecision = prec
	}
}

// OtoWithDirectory specifies the path to the directory where the oto.ini file is to put
// into the [OtoEntry.Directory] field when reading the oto.ini file.
func OtoWithDirectory(dir string) OtoOption {
	return func(cfg *otoConfig) {
		cfg.dir = dir
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

		filename, _values, ok := strings.Cut(line, "=")
		if !ok {
			return oto, fmt.Errorf("invalid oto entry: %s", strconv.Quote(line))
		}
		values := strings.Split(_values, ",")

		// filename=alias,offset,consonant,cutoff,preutter,overlap
		// filename and alias are strings, the rest are floats
		if len(values) != 6 {
			return oto, fmt.Errorf("invalid oto entry for %s, expected 6 values, got %d", strconv.Quote(filename), len(values))
		}

		alias := values[0]

		if values[1] == "" {
			values[1] = "0"
		}
		offset, err := strconv.ParseFloat(values[1], 64)
		if err != nil {
			return oto, fmt.Errorf("failed to parse offset value for %s: %w", strconv.Quote(filename), err)
		}

		if values[2] == "" {
			values[2] = "0"
		}
		consonant, err := strconv.ParseFloat(values[2], 64)
		if err != nil {
			return oto, fmt.Errorf("failed to parse consonant value for %s: %w", strconv.Quote(filename), err)
		}

		if values[3] == "" {
			values[3] = "0"
		}
		cutoff, err := strconv.ParseFloat(values[3], 64)
		if err != nil {
			return oto, fmt.Errorf("failed to parse cutoff value for %s: %w", strconv.Quote(filename), err)
		}

		if values[4] == "" {
			values[4] = "0"
		}
		preutter, err := strconv.ParseFloat(values[4], 64)
		if err != nil {
			return oto, fmt.Errorf("failed to parse preutterance value for %s: %w", strconv.Quote(filename), err)
		}

		if values[5] == "" {
			values[5] = "0"
		}
		overlap, err := strconv.ParseFloat(values[5], 64)
		if err != nil {
			return oto, fmt.Errorf("failed to parse overlap value for %s: %w", strconv.Quote(filename), err)
		}

		oto = append(oto, OtoEntry{
			Filename:     filename,
			Directory:    cfg.dir,
			Alias:        alias,
			Offset:       offset,
			Consonant:    consonant,
			Cutoff:       cutoff,
			Preutterance: preutter,
			Overlap:      overlap,
		})
	}
	return oto, scan.Err()
}

//TODO: optimize with binary search and lookup table

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
		encoding:       encoding.Nop,
		floatPrecision: -1,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	newWriter := w
	if cfg.encoding != encoding.Nop {
		newWriter = transform.NewWriter(w, cfg.encoding.NewEncoder())
	}
	var buf bytes.Buffer
	buf.Grow(256)                   // preallocate some space
	floatTmp := make([]byte, 0, 64) // scratch buffer for floats

	for _, entry := range o {
		// filename=alias,offset,consonant,cutoff,preutter,overlap
		// filename and alias are strings, the rest are floats
		buf.Reset()

		buf.WriteString(entry.Filename)
		buf.WriteByte('=')
		buf.WriteString(entry.Alias)
		buf.WriteByte(',')
		buf.Write(strconv.AppendFloat(floatTmp[:0], entry.Offset, 'f', cfg.floatPrecision, 64))
		buf.WriteByte(',')
		buf.Write(strconv.AppendFloat(floatTmp[:0], entry.Consonant, 'f', cfg.floatPrecision, 64))
		buf.WriteByte(',')
		buf.Write(strconv.AppendFloat(floatTmp[:0], entry.Cutoff, 'f', cfg.floatPrecision, 64))
		buf.WriteByte(',')
		buf.Write(strconv.AppendFloat(floatTmp[:0], entry.Preutterance, 'f', cfg.floatPrecision, 64))
		buf.WriteByte(',')
		buf.Write(strconv.AppendFloat(floatTmp[:0], entry.Overlap, 'f', cfg.floatPrecision, 64))
		buf.WriteByte(10) // newline

		if _, err := buf.WriteTo(newWriter); err != nil {
			return fmt.Errorf("failed to write oto entry for %s: %w", strconv.Quote(entry.Filename), err)
		}
	}
	return nil
}
