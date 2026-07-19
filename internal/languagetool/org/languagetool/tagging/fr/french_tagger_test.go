package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestFrenchTagger(t *testing.T) {
	wt := tagging.MapWordTagger{"chien": {tagging.NewTaggedWord("chien", "N m s")}}
	got := NewFrenchTagger(wt).Tag([]string{"chien", "xyz"})
	require.Len(t, got, 2)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
}

func TestFrenchTagger_ApostropheChunkTags(t *testing.T) {
	wt := tagging.MapWordTagger{"l'eau": {tagging.NewTaggedWord("eau", "N")}}
	tagger := NewFrenchTagger(wt)
	// typewriter
	got := tagger.Tag([]string{"l'eau"})
	require.Contains(t, got[0].GetChunkTags(), "containsTypewriterApostrophe")
	// typographic overwrites typewriter list per Java
	got2 := tagger.Tag([]string{"l’eau"})
	require.Contains(t, got2[0].GetChunkTags(), "containsTypographicApostrophe")
	require.NotContains(t, got2[0].GetChunkTags(), "containsTypewriterApostrophe")
}
