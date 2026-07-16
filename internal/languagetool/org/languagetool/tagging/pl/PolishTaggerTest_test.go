package pl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestPolishTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"dom": {tagging.NewTaggedWord("dom", "N")},
	}
	tagger := NewPolishTagger(wt)
	got := tagger.Tag([]string{"dom", "xyz"})
	require.Len(t, got, 2)
	require.NotEmpty(t, got[0].GetReadings())
	// unknown word still yields a reading
	require.NotEmpty(t, got[1].GetReadings())
}

func TestPolishTagger_DictionaryPath(t *testing.T) {
	tagger := NewPolishTagger(nil)
	require.NotEmpty(t, tagger.GetDictionaryPath())
}
