package gotau

import (
	"encoding/binary"
	"io"
	"io/fs"

	"github.com/SladkyCitron/gotau/cache"
	"github.com/SladkyCitron/gotau/resample"
)

const keyVersion uint64 = 1

func (s *Synth) getKeyFunc(cfg resample.ResampleConfig, path string, fileinfo fs.FileInfo) cache.KeyFunc {
	mtime, err := fileinfo.ModTime().MarshalBinary()
	if err != nil {
		mtime = []byte{}
	}

	return func(w io.Writer) {
		_, _ = w.Write([]byte("gotau-resample"))
		_ = binary.Write(w, binary.LittleEndian, keyVersion)
		_, _ = w.Write([]byte(s.res.ID()))
		_, _ = w.Write([]byte(path))
		_ = binary.Write(w, binary.LittleEndian, fileinfo.Size())
		_, _ = w.Write(mtime)
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
