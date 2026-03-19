package phonemizer_test

import (
	"slices"
	"testing"

	"github.com/SladkyCitron/gotau/phonemizer"
	"github.com/SladkyCitron/gotau/voicebank"
	"github.com/stretchr/testify/assert"
)

func TestJapaneseVCV(t *testing.T) {
	p := &phonemizer.JapaneseVCV{PrefixMap: voicebank.PrefixMap{60: voicebank.Prefix{"pre", "suf"}}}
	got := slices.Collect(p.Resolve(phonemizer.ResolveConfig{
		Lyric: "か ",
		Note:  60,
	}))

	want := []string{
		"pre- か suf", // prefix.map
		"- か",        // trimmed
		"- か ",       // raw
	}

	assert.Equal(t, want, got)
}
