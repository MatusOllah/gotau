package gotau

import (
	"io"

	"github.com/SladkyCitron/gotau/cache"
	"github.com/SladkyCitron/gotau/resample"
)

func (s *Synth) getKeyFunc(cfg resample.ResampleConfig) cache.KeyFunc {
	return func(w io.Writer) error {
		return nil
	}
}
