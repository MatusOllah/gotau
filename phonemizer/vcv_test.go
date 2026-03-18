package phonemizer_test

import (
	"slices"
	"testing"

	"github.com/SladkyCitron/gotau/phonemizer"
	"github.com/SladkyCitron/gotau/voicebank"
	"github.com/stretchr/testify/assert"
)

func TestVCV(t *testing.T) {
	p := &phonemizer.VCV{PrefixMap: voicebank.PrefixMap{60: voicebank.Prefix{"pre", "suf"}}}
	got := slices.Collect(p.Resolve(phonemizer.ResolveConfig{
		Lyric: "ka ",
		Note:  60,
	}))

	want := []string{
		"pre- ka suf", // prefix.map
		"- ka",        // trimmed
		"- ka ",       // raw
	}

	assert.Equal(t, want, got)
}
