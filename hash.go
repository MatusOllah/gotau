package gotau

import (
	"encoding/binary"
	"io"

	"github.com/SladkyCitron/gotau/cache"
	"github.com/SladkyCitron/gotau/resample"
)

func (s *Synth) getKeyFunc(cfg resample.ResampleConfig) cache.KeyFunc {
	return func(w io.Writer) {
		_, _ = w.Write([]byte("gotau-resample"))
		_, _ = w.Write([]byte(s.res.ID()))
		_ = binary.Write(w, binary.LittleEndian, uint64(s.vbFileBuf.Len()))
		_, _ = w.Write(s.vbFileBuf.Bytes())
		_, _ = w.Write([]byte{byte(cfg.Pitch)})
		_ = binary.Write(w, binary.LittleEndian, cfg.Velocity)
		_, _ = w.Write([]byte(cfg.Flags))
		_ = binary.Write(w, binary.LittleEndian, cfg.Offset)
		_ = binary.Write(w, binary.LittleEndian, cfg.Length)
		_ = binary.Write(w, binary.LittleEndian, cfg.Consonant)
		_ = binary.Write(w, binary.LittleEndian, cfg.Cutoff)
		_ = binary.Write(w, binary.LittleEndian, cfg.Intensity)
		_ = binary.Write(w, binary.LittleEndian, cfg.Modulation)
		_ = binary.Write(w, binary.LittleEndian, cfg.Tempo)
		_ = binary.Write(w, binary.LittleEndian, uint64(cfg.Resolution))
		_ = binary.Write(w, binary.LittleEndian, uint64(len(cfg.PitchBend)))
		for _, pt := range cfg.PitchBend {
			_ = binary.Write(w, binary.LittleEndian, uint64(pt.X))
			_ = binary.Write(w, binary.LittleEndian, pt.Y)
			_, _ = w.Write([]byte{byte(pt.Interp)})
		}
	}
}
