package sequence

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// OptionFlagValue is a special value used to indicate that an option flag is set.
const OptionFlagValue = math.MaxInt64

// we don't need helper methods like Get, Set, etc. because it's just a map

// Flags is a collection of resampler flags. Each flag is a string key with an integer value.
// Option flags can be represented with [OptionFlagValue], and the presence of the key itself indicates the flag is set.
type Flags map[string]int

// Has checks whether a given flag is set.
func (f Flags) Has(key string) bool {
	_, ok := f[key]
	return ok
}

// String returns a canonical UTAU string representation of the flags.
//
// The format for each flag is:
//
//	<key><value>
//
// for numeric flags (e.g. "B50", "g-10"), and just
//
//	<key>
//
// for option flags (e.g. "G").
//
// The flags are sorted alphabetically by key to ensure a deterministic output.
func (f Flags) String() string {
	if len(f) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.Grow(64)                     // allocate some space
	valueTmp := make([]byte, 0, 16) // scratch buffer for values

	keys := slices.Sorted(maps.Keys(f))
	for _, key := range keys {
		_, _ = sb.WriteString(key)
		value := f[key]
		if value != OptionFlagValue {
			_, _ = sb.Write(strconv.AppendInt(valueTmp[:0], int64(value), 10))
		}
	}

	return sb.String()
}

// AppendBinary implements the [encoding.BinaryAppender] interface.
//
// It's mostly used for hashing and caching purposes.
func (f Flags) AppendBinary(b []byte) ([]byte, error) {
	if len(f) == 0 {
		return b, nil
	}

	keys := slices.Sorted(maps.Keys(f))

	tmp := make([]byte, binary.MaxVarintLen64)

	// number of flags
	n := binary.PutUvarint(tmp, uint64(len(keys)))
	b = append(b, tmp[:n]...)

	for _, key := range keys {
		// key length
		n = binary.PutUvarint(tmp, uint64(len(key)))
		b = append(b, tmp[:n]...)
		// key bytes
		b = append(b, key...)

		// value
		n = binary.PutVarint(tmp, int64(f[key]))
		b = append(b, tmp[:n]...)
	}

	return b, nil
}

// MarshalBinary implements the [encoding.BinaryMarshaler] interface.
//
// It's mostly used for hashing and caching purposes.
func (f Flags) MarshalBinary() ([]byte, error) {
	return f.AppendBinary(make([]byte, 0, 64))
}

// https://github.com/stakira/OpenUtau/blob/master/OpenUtau.Core/Classic/Flags/UstFlagParser.cs

// ParseFlags parses a UTAU resampler flag string into a [Flags] map.
func ParseFlags(s string) (Flags, error) {
	flags := make(Flags)
	if s == "" || s == "?" {
		return flags, nil
	}
	var keyBuf bytes.Buffer
	keyBuf.Grow(8)
	var valueBuf bytes.Buffer
	valueBuf.Grow(8)
	wasDigit := false
	i := 0
	for i <= len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if i == len(s) || unicode.IsLetter(r) && wasDigit {
			key := keyBuf.String()
			value, err := strconv.Atoi(valueBuf.String())
			if err != nil {
				return nil, fmt.Errorf("sequence ParseFlags: invalid flag value for key %q: %w", key, err)
			}
			if key != "" {
				flags[key] = value
			}
			keyBuf.Reset()
			valueBuf.Reset()
			wasDigit = false
		}
		if i == len(s) {
			break
		}
		if r == '-' || r == '+' || unicode.IsDigit(r) {
			_, _ = valueBuf.WriteRune(r)
			wasDigit = true
		} else if unicode.IsLetter(r) {
			if keyBuf.Len() == 0 && isOptionFlag(r) {
				flags[string(r)] = OptionFlagValue
				i += size
				continue
			}
			_, _ = keyBuf.WriteRune(r)
			wasDigit = false
		} else {
			return nil, fmt.Errorf("sequence ParseFlags: invalid character in flags: %q", r)
		}
		i += size
	}
	return flags, nil
}

func isOptionFlag(r rune) bool {
	return strings.ContainsRune("GNeu", r)
}
