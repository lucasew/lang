package sk

// Twin of SlovakTaggerTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestSlovakTagger_Dictionary(t *testing.T) {
	wt := tagging.MapWordTagger{"dom": {tagging.NewTaggedWord("dom", "SSis1")}}
	tagger := NewSlovakTagger(wt)
	require.Equal(t, SlovakDictPath, tagger.GetDictionaryPath())
	require.Len(t, tagger.TagWord("dom"), 1)
}

func TestSlovakTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"toto": {tagging.NewTaggedWord("toto", "PFns1")},
		"test": {tagging.NewTaggedWord("test", "SSis1")},
	}
	got := NewSlovakTagger(wt).Tag([]string{"Toto", "test", "xyz"})
	require.Len(t, got, 3)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.Nil(t, got[2].GetReadings()[0].GetPOSTag())
}
