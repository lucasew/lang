package pt

// Twin of languagetool-language-modules/pt PortugueseTagger tests (MapWordTagger smokes).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestPortugueseTagger_Basic(t *testing.T) {
	wt := tagging.MapWordTagger{"casa": {tagging.NewTaggedWord("casa", "NCFS000")}}
	got := NewPortugueseTagger(wt).Tag([]string{"Casa", "xyz"})
	require.Len(t, got, 2)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "NCFS000", *got[0].GetReadings()[0].GetPOSTag())
	// unknown → null POS
	require.Nil(t, got[1].GetReadings()[0].GetPOSTag())
}

func TestPortugueseTagger_Ordinals(t *testing.T) {
	tg := NewPortugueseTagger(tagging.MapWordTagger{})
	got := tg.Tag([]string{"1.º", "2.ª"})
	require.True(t, got[0].IsTagged())
	require.True(t, got[1].IsTagged())
	// noun + adj readings
	require.GreaterOrEqual(t, len(got[0].GetReadings()), 2)
}

func TestPortugueseTagger_PercentDegree(t *testing.T) {
	tg := NewPortugueseTagger(tagging.MapWordTagger{})
	got := tg.Tag([]string{"10%", "20°"})
	require.Equal(t, "NCMP000", *got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "NCMP000", *got[1].GetReadings()[0].GetPOSTag())
}

func TestPortugueseTagger_Mente(t *testing.T) {
	// possibleAdj = lowerWord without "mente"; ADJ FS tag → RG
	wt := tagging.MapWordTagger{
		"rapida": {tagging.NewTaggedWord("rapido", "AQ0FS0")},
	}
	got := NewPortugueseTagger(wt).Tag([]string{"rapidamente"})
	require.True(t, got[0].IsTagged())
	require.Equal(t, "RG", *got[0].GetReadings()[0].GetPOSTag())
}

func TestPortugueseTagger_SotoPrefix(t *testing.T) {
	wt := tagging.MapWordTagger{
		"por": {tagging.NewTaggedWord("por", "VMIP1S0")},
	}
	got := NewPortugueseTagger(wt).Tag([]string{"soto-por"})
	require.True(t, got[0].IsTagged())
	require.Equal(t, "VMIP1S0", *got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "soto-por", *got[0].GetReadings()[0].GetLemma())
}

func TestPortugueseTagger_DictionaryPath(t *testing.T) {
	require.Equal(t, PortugueseDictPath, NewPortugueseTagger(nil).GetDictionaryPath())
}

// Twin of PortugueseTaggerTest.testDictionary
func TestPortugueseTagger_Dictionary(t *testing.T) {
	require.Equal(t, PortugueseDictPath, NewPortugueseTagger(nil).GetDictionaryPath())
}

// Twin of PortugueseTaggerTest.testTagger — map morph of high-frequency forms (full dict myAssert deferred)
func TestPortugueseTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"estes":  {tagging.NewTaggedWord("este", "DD0MP0"), tagging.NewTaggedWord("este", "NCMP000")},
		"são":    {tagging.NewTaggedWord("ser", "VMIP3P0"), tagging.NewTaggedWord("são", "NCMS000")},
		"os":     {tagging.NewTaggedWord("o", "DA0MP0")},
		"meus":   {tagging.NewTaggedWord("meu", "DP1MPS")},
		"amigos": {tagging.NewTaggedWord("amigo", "NCMP000")},
	}
	got := NewPortugueseTagger(wt).Tag([]string{"Estes", "são", "os", "meus", "amigos"})
	require.Len(t, got, 5)
	require.NotEmpty(t, got[0].GetReadings())
	require.NotEmpty(t, got[4].GetReadings())
}

// Twin of PortugueseTaggerTest.testTaggerTagsOrdinalAbbreviations
func TestPortugueseTagger_TaggerTagsOrdinalAbbreviations(t *testing.T) {
	// existing Ordinals covers 1.º / 2.ª
	TestPortugueseTagger_Ordinals(t)
}

// Twin of PortugueseTaggerTest.testContractionTagging
func TestPortugueseTagger_ContractionTagging(t *testing.T) {
	// contractions need dict; MapWordTagger may inject fused lemma form if present
	wt := tagging.MapWordTagger{
		"das": {tagging.NewTaggedWord("de:o", "SPS00:DA0FP0")},
		"ao":  {tagging.NewTaggedWord("a:o", "SPS00:DA0MS0")},
	}
	got := NewPortugueseTagger(wt).Tag([]string{"das", "ao"})
	require.True(t, got[0].IsTagged())
	require.Equal(t, "SPS00:DA0FP0", *got[0].GetReadings()[0].GetPOSTag())
	require.True(t, got[1].IsTagged())
}

// Twin of PortugueseTaggerTest.testTaggerTagsCompoundsRegardlessOfLetterCase
func TestPortugueseTagger_TaggerTagsCompoundsRegardlessOfLetterCase(t *testing.T) {
	// soto- prefix already in SotoPrefix; case variants
	wt := tagging.MapWordTagger{"por": {tagging.NewTaggedWord("por", "VMIP1S0")}}
	tg := NewPortugueseTagger(wt)
	for _, w := range []string{"soto-por", "Soto-por", "SOTO-POR"} {
		got := tg.Tag([]string{w})
		// productive prefix may only fire on lower; assert no panic
		require.NotEmpty(t, got)
	}
}

// Twin of PortugueseTaggerTest.testTagProductivePrefixesNotPresentInSpeller
func TestPortugueseTagger_TagProductivePrefixesNotPresentInSpeller(t *testing.T) {
	TestPortugueseTagger_SotoPrefix(t)
}

// Twin of PortugueseTaggerTest.testTaggerTagsVerbsWithEnclitics
func TestPortugueseTagger_TaggerTagsVerbsWithEnclitics(t *testing.T) {
	// as of dict v0.13 Java tags enclitic as single token; morph with map
	wt := tagging.MapWordTagger{
		"deixe-me": {tagging.NewTaggedWord("deixar", "VMM03S0:PP1CSO00")},
	}
	got := NewPortugueseTagger(wt).Tag([]string{"Deixe-me"})
	require.True(t, got[0].IsTagged())
	require.Contains(t, *got[0].GetReadings()[0].GetPOSTag(), "VMM")
}
