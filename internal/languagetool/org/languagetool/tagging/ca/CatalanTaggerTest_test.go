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
	require.Equal(t, "/ca/ca-ES.dict", CatalanDictPath)
	got := tagger.TagWordExact("casa")
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

func TestCatalanTagger_TypographicApostrophe(t *testing.T) {
	tagger := NewCatalanTagger(tagging.MapWordTagger{
		"l'home": {tagging.NewTaggedWord("home", "NCMS000")},
	})
	got := tagger.Tag([]string{"l’home"})
	require.Len(t, got, 1)
	require.True(t, got[0].HasTypographicApostrophe())
	// Java replaces ’ → ' on originalWord used as surface
	require.Equal(t, "l'home", got[0].GetToken())
	require.True(t, got[0].IsTagged())
}

func TestCatalanTagger_AllUpper(t *testing.T) {
	wt := tagging.MapWordTagger{
		"França": {tagging.NewTaggedWord("França", "NPFS000")},
	}
	got := NewCatalanTagger(wt).Tag([]string{"FRANÇA"})
	require.True(t, got[0].IsTagged())
	require.Equal(t, "NPFS000", *got[0].GetReadings()[0].GetPOSTag())
}

func TestCatalanTagger_Ment(t *testing.T) {
	wt := tagging.MapWordTagger{
		"rapida": {tagging.NewTaggedWord("rapid", "AQ0FS0")},
	}
	got := NewCatalanTagger(wt).Tag([]string{"rapidament"})
	require.True(t, got[0].IsTagged())
	require.Equal(t, "RG", *got[0].GetReadings()[0].GetPOSTag())
}

func TestCatalanTagger_FilterValencianZeroPrefix(t *testing.T) {
	wt := tagging.MapWordTagger{
		"casa": {tagging.NewTaggedWord("casa", "0NCFS000")},
	}
	// central: drop 0* tags → untagged
	got := NewCatalanTagger(wt).Tag([]string{"casa"})
	require.False(t, got[0].IsTagged())
	// valencian: strip leading 0
	gotV := NewCatalanTaggerValencian(wt).Tag([]string{"casa"})
	require.True(t, gotV[0].IsTagged())
	require.Equal(t, "NCFS000", *gotV[0].GetReadings()[0].GetPOSTag())
}

func TestCatalanTagger_DictionaryPath(t *testing.T) {
	require.Equal(t, "/ca/ca-ES.dict", NewCatalanTagger(nil).GetDictionaryPath())
}
