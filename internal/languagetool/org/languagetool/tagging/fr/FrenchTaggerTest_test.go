package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestFrenchTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"maison": {tagging.NewTaggedWord("maison", "N")},
	}
	tagger := NewFrenchTagger(wt)
	got := tagger.Tag([]string{"maison", "xyz"})
	require.Len(t, got, 2)
	require.NotEmpty(t, got[0].GetReadings())
	// unknown word still yields a reading
	require.NotEmpty(t, got[1].GetReadings())
}

func TestFrenchTagger_DictionaryPath(t *testing.T) {
	tagger := NewFrenchTagger(nil)
	require.NotEmpty(t, tagger.GetDictionaryPath())
}
