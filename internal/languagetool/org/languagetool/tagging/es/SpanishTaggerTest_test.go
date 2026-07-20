package es

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestSpanishTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"casa": {tagging.NewTaggedWord("casa", "NCFS000")},
	}
	tagger := NewSpanishTagger(wt)
	got := tagger.Tag([]string{"Casa", "xyz"})
	require.Len(t, got, 2)
	require.Equal(t, "NCFS000", *got[0].GetReadings()[0].GetPOSTag())
	require.Nil(t, got[1].GetReadings()[0].GetPOSTag())
}

func TestSpanishTagger_AllUpper(t *testing.T) {
	wt := tagging.MapWordTagger{
		"Francia": {tagging.NewTaggedWord("Francia", "NPFS000")},
	}
	got := NewSpanishTagger(wt).Tag([]string{"FRANCIA"})
	require.True(t, got[0].IsTagged())
	require.Equal(t, "NPFS000", *got[0].GetReadings()[0].GetPOSTag())
}

func TestSpanishTagger_Mente(t *testing.T) {
	wt := tagging.MapWordTagger{
		"rapida": {tagging.NewTaggedWord("rapido", "AQ0FS0")},
	}
	got := NewSpanishTagger(wt).Tag([]string{"rapidamente"})
	require.True(t, got[0].IsTagged())
	require.Equal(t, "RG", *got[0].GetReadings()[0].GetPOSTag())
}

func TestSpanishTagger_AutoVerb(t *testing.T) {
	wt := tagging.MapWordTagger{
		"destruir": {tagging.NewTaggedWord("destruir", "VMN0000")},
	}
	got := NewSpanishTagger(wt).Tag([]string{"autodestruir"})
	require.True(t, got[0].IsTagged())
	require.Equal(t, "VMN0000", *got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "autodestruir", *got[0].GetReadings()[0].GetLemma())
}

func TestSpanishTagger_DictionaryPath(t *testing.T) {
	tagger := NewSpanishTagger(nil)
	require.Equal(t, SpanishDictPath, tagger.GetDictionaryPath())
}

func TestSpanishTagger_TypographicApostrophe(t *testing.T) {
	wt := tagging.MapWordTagger{"d'": {tagging.NewTaggedWord("de", "SPS00")}}
	tagger := NewSpanishTagger(wt)
	got := tagger.Tag([]string{"d’"})
	require.Len(t, got, 1)
	require.True(t, got[0].HasTypographicApostrophe())
}
