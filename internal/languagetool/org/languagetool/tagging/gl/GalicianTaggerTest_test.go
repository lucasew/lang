package gl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestGalicianTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"casa": {tagging.NewTaggedWord("casa", "N")},
	}
	tagger := NewGalicianTagger(wt)
	got := tagger.Tag([]string{"casa", "xyz"})
	require.Len(t, got, 2)
	require.NotEmpty(t, got[0].GetReadings())
	// unknown word still yields a reading
	require.NotEmpty(t, got[1].GetReadings())
}

func TestGalicianTagger_DictionaryPath(t *testing.T) {
	tagger := NewGalicianTagger(nil)
	require.NotEmpty(t, tagger.GetDictionaryPath())
}
