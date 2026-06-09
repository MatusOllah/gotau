package pitch

import (
	"math"
	"strconv"
	"strings"
)

// EncodeResamplerPitchBendString encodes a pitch bend curve into the UTAU resampler pitch bend string format.
func EncodeResamplerPitchBendString(x []float64) string {
	if len(x) == 0 {
		return "AA"
	}

	last := int16(math.MinInt16)
	run := 0 // run length

	var buf strings.Builder
	buf.Grow(len(x) * 2)         // allocate some space
	runTmp := make([]byte, 0, 8) // scratch buffer for run length

	for i := range x {
		num := int16(math.Round(x[i]))
		if num == last {
			run++
			continue
		}

		// flush run
		if run > 0 {
			buf.WriteByte('#')
			buf.Write(strconv.AppendInt(runTmp[:0], int64(run), 10))
			buf.WriteByte('#')
			run = 0
		}

		writeInt12(&buf, num)
		last = num
	}

	// flush remaining run
	if run > 0 {
		buf.WriteByte('#')
		buf.Write(strconv.AppendInt(runTmp[:0], int64(run), 10))
		buf.WriteByte('#')
	}

	return buf.String()
}

const b64 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

func writeInt12(buf *strings.Builder, v int16) {
	if v == 0 {
		if _, err := buf.WriteString("AA"); err != nil {
			panic(err)
		}
		return
	}
	if v < -2048 {
		v = -2048
	}
	if v > 2047 {
		v = 2047
	}
	if v < 0 {
		v += 4096
	}

	hi := (v >> 6) & 0x3f
	lo := v & 0x3f

	if err := buf.WriteByte(b64[hi]); err != nil {
		panic(err)
	}
	if err := buf.WriteByte(b64[lo]); err != nil {
		panic(err)
	}
}
