package it

// Twin of ItalianTaggerTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestItalianTagger_Dictionary(t *testing.T) {
	wt := tagging.MapWordTagger{"casa": {tagging.NewTaggedWord("casa", "S")}}
	tagger := NewItalianTagger(wt)
	require.Equal(t, ItalianDictPath, tagger.GetDictionaryPath())
	require.Len(t, tagger.TagWord("casa"), 1)
}

func TestItalianTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"cane": {tagging.NewTaggedWord("cane", "S")},
		"bello": {tagging.NewTaggedWord("bello", "A")},
	}
	got := NewItalianTagger(wt).Tag([]string{"Cane", "bello", "xyz"})
	require.Len(t, got, 3)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.Nil(t, got[2].GetReadings()[0].GetPOSTag())
}
