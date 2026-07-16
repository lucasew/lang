package sr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestEkavianTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{"raditi": {tagging.NewTaggedWord("raditi", "V")}}
	tagger := NewEkavianTagger(wt)
	got := tagger.TagWord("raditi")
	require.Len(t, got, 1)
	require.Equal(t, "V", got[0].GetPosTag())
}

func TestEkavianTagger_Dictionary(t *testing.T) {
	require.Equal(t, EkavianDictionaryPath, NewEkavianTagger(nil).GetDictionaryPath())
}
