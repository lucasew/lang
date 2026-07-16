package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestEnglishTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"house": {tagging.NewTaggedWord("house", "N")},
	}
	tagger := NewEnglishTagger(wt)
	got := tagger.Tag([]string{"house", "xyz"})
	require.Len(t, got, 2)
	require.NotEmpty(t, got[0].GetReadings())
	// unknown word still yields a reading
	require.NotEmpty(t, got[1].GetReadings())
}

func TestEnglishTagger_DictionaryPath(t *testing.T) {
	tagger := NewEnglishTagger(nil)
	require.NotEmpty(t, tagger.GetDictionaryPath())
}
