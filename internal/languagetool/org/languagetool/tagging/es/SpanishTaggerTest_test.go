package es

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestSpanishTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"casa": {tagging.NewTaggedWord("casa", "N")},
	}
	tagger := NewSpanishTagger(wt)
	got := tagger.Tag([]string{"casa", "xyz"})
	require.Len(t, got, 2)
	require.NotEmpty(t, got[0].GetReadings())
	// unknown word still yields a reading
	require.NotEmpty(t, got[1].GetReadings())
}

func TestSpanishTagger_DictionaryPath(t *testing.T) {
	tagger := NewSpanishTagger(nil)
	require.NotEmpty(t, tagger.GetDictionaryPath())
}

func TestSpanishTagger_TypographicApostrophe(t *testing.T) {
	wt := tagging.MapWordTagger{"d'": {tagging.NewTaggedWord("de", "SPS00")}}
	tagger := NewSpanishTagger(wt)
	got := tagger.Tag([]string{"d’"})
	require.Len(t, got, 1)
	require.True(t, got[0].HasTypographicApostrophe())
}
