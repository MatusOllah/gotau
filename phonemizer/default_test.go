package phonemizer_test

import (
	"slices"
	"testing"

	"github.com/SladkyCitron/gotau/phonemizer"
	"github.com/stretchr/testify/assert"
)

func TestDefault(t *testing.T) {
	want := "a"

	p := &phonemizer.Default{}
	got := slices.Collect(p.Resolve(phonemizer.ResolveConfig{Lyric: want}))

	wantSlice := []string{want}
	assert.Equal(t, wantSlice, got)
}
