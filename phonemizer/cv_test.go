package phonemizer_test

import (
	"slices"
	"testing"

	"github.com/SladkyCitron/gotau/phonemizer"
	"github.com/SladkyCitron/gotau/voicebank"
	"github.com/stretchr/testify/assert"
)

func TestCV(t *testing.T) {
	p := &phonemizer.CV{PrefixMap: voicebank.PrefixMap{60: voicebank.Prefix{"pre", "suf"}}}
	got := slices.Collect(p.Resolve(phonemizer.ResolveConfig{
		Lyric: " a ",
		Note:  60,
	}))

	want := []string{
		"pre a suf", // prefix.map
		"a",         // trimmed
		" a ",       // raw
	}

	assert.Equal(t, want, got)
}
