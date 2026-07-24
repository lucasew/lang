package crh

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestCrimeanTatarTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"ev": {tagging.NewTaggedWord("ev", "N")},
	}
	tagger := NewCrimeanTatarTagger(wt)
	got := tagger.TagWord("ev")
	require.Len(t, got, 1)
	require.Equal(t, "N", got[0].GetPosTag())
	require.Empty(t, tagger.TagWord("xyz"))
}

func TestCrimeanTatarTagger_Dictionary(t *testing.T) {
	require.Equal(t, CrimeanTatarTaggerDictPath, NewCrimeanTatarTagger(nil).GetDictionaryPath())
}
