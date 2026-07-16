package nl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestDutchTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"huis": {tagging.NewTaggedWord("huis", "N")},
	}
	tagger := NewDutchTagger(wt)
	got := tagger.Tag([]string{"huis", "xyz"})
	require.Len(t, got, 2)
	require.NotEmpty(t, got[0].GetReadings())
	// unknown word still yields a reading
	require.NotEmpty(t, got[1].GetReadings())
}

func TestDutchTagger_DictionaryPath(t *testing.T) {
	tagger := NewDutchTagger(nil)
	require.NotEmpty(t, tagger.GetDictionaryPath())
}
