package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/tagging/ca/CatalanTaggerTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

// Port of CatalanTaggerTest.testDictionary — full dict deferred; MapWordTagger smoke
func TestCatalanTagger_Dictionary(t *testing.T) {
	wt := tagging.MapWordTagger{
		"casa": {tagging.NewTaggedWord("casa", "NCFS000")},
	}
	tagger := NewCatalanTagger(wt)
	require.NotNil(t, tagger)
	require.Equal(t, CatalanDictPath, tagger.GetDictionaryPath())
	got := tagger.TagWord("casa")
	require.Len(t, got, 1)
	require.Equal(t, "NCFS000", got[0].PosTag)
}

// Port of CatalanTaggerTest.testTagger
func TestCatalanTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"casa":  {tagging.NewTaggedWord("casa", "NCFS000")},
		"blava": {tagging.NewTaggedWord("blau", "AQ0FS0")},
	}
	got := NewCatalanTagger(wt).Tag([]string{"Casa", "blava", "xyz"})
	require.Len(t, got, 3)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.NotNil(t, got[1].GetReadings()[0].GetPOSTag())
	require.Nil(t, got[2].GetReadings()[0].GetPOSTag())
}
