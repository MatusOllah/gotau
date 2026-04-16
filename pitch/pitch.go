package pitch

import (
	"math"
	"strconv"
	"strings"

	"github.com/SladkyCitron/gotau/internal/timeutil"
	"github.com/SladkyCitron/gotau/sequence"
	"gitlab.com/gomidi/midi/v2"
)

type int12 int16

// EncodeResamplerPitchBendString encodes a pitch bend curve into the UTAU resampler pitch bend string format.
func EncodeResamplerPitchBendString(curve sequence.Curve, note midi.Note, durationSec float64, tempo float64, tpqn int) string {
	if len(curve) == 0 {
		return "AA"
	}

	durationMs := durationSec * 1000
	last := int12(math.MinInt16)
	run := 0 // run length

	var buf strings.Builder
	buf.Grow(int(math.Round(durationMs/5)) * 2) // allocate some space
	runTmp := make([]byte, 0, 8)                // scratch buffer for run length

	for t := float64(0); t <= durationMs; t += 5 {
		pitch := curve.At(timeutil.SecondsToTicks(t/1000, tpqn, tempo))
		diffCents := int12(math.Round(pitch - float64(note)*100))
		if diffCents == last {
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

		writeInt12(&buf, diffCents)
		last = diffCents
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

func writeInt12(buf *strings.Builder, v int12) {
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
