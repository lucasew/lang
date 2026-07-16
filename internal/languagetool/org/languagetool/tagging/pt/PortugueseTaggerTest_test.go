package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/tagging/pt/PortugueseTaggerTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

// Port of PortugueseTaggerTest.testDictionary
func TestPortugueseTagger_Dictionary(t *testing.T) {
	wt := tagging.MapWordTagger{
		"casa": {tagging.NewTaggedWord("casa", "NCFS000")},
	}
	tagger := NewPortugueseTagger(wt)
	require.Equal(t, PortugueseDictPath, tagger.GetDictionaryPath())
	got := tagger.TagWord("casa")
	require.Len(t, got, 1)
}

// Port of PortugueseTaggerTest.testTagger
func TestPortugueseTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"isto":  {tagging.NewTaggedWord("isto", "PD0NS000")},
		"frase": {tagging.NewTaggedWord("frase", "NCFS000")},
	}
	got := NewPortugueseTagger(wt).Tag([]string{"Isto", "frase", "xyz"})
	require.Len(t, got, 3)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.NotNil(t, got[1].GetReadings()[0].GetPOSTag())
	require.Nil(t, got[2].GetReadings()[0].GetPOSTag())
}

// Remaining Java cases need real dict / clitics — soft skip until resources land.
func TestPortugueseTagger_TaggerTagsOrdinalAbbreviations(t *testing.T) {
	t.Skip("unimplemented: needs full Portuguese dict for ordinal abbreviations")
}
func TestPortugueseTagger_ContractionTagging(t *testing.T) {
	t.Skip("unimplemented: contraction tagging needs dict")
}
func TestPortugueseTagger_TaggerTagsCompoundsRegardlessOfLetterCase(t *testing.T) {
	t.Skip("unimplemented: compound tagging needs dict")
}
func TestPortugueseTagger_TagProductivePrefixesNotPresentInSpeller(t *testing.T) {
	t.Skip("unimplemented: productive prefixes need dict")
}
func TestPortugueseTagger_TaggerTagsVerbsWithEnclitics(t *testing.T) {
	t.Skip("unimplemented: enclitics need dict")
}
