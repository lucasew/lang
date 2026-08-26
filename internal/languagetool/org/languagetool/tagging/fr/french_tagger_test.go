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
	// Java surface becomes typewriter after replace
	require.Equal(t, "l'eau", got2[0].GetToken())
}

func TestFrenchTagger_CapitalizedAndAllUpper(t *testing.T) {
	wt := tagging.MapWordTagger{
		"france": {tagging.NewTaggedWord("France", "N m s")},
		"France": {tagging.NewTaggedWord("France", "N m s prop")},
	}
	tagger := NewFrenchTagger(wt)
	// Capitalized merges lower
	got := tagger.Tag([]string{"France"})
	require.True(t, got[0].IsTagged())
	// ALL UPPER tries ConvertToTitleCase when empty of exact+lower
	got2 := tagger.Tag([]string{"FRANCE"})
	require.True(t, got2[0].IsTagged())
}

func TestFrenchTagger_OeLigature(t *testing.T) {
	wt := tagging.MapWordTagger{
		"cœur": {tagging.NewTaggedWord("cœur", "N m s")},
	}
	got := NewFrenchTagger(wt).Tag([]string{"coeur"})
	require.True(t, got[0].IsTagged())
	require.Equal(t, "N m s", *got[0].GetReadings()[0].GetPOSTag())
}

func TestFrenchTagger_VerbPrefix(t *testing.T) {
	wt := tagging.MapWordTagger{
		// need two vowels in remainder for PREFIXES_FOR_VERBS
		"définir": {tagging.NewTaggedWord("définir", "V inf")},
	}
	got := NewFrenchTagger(wt).Tag([]string{"redéfinir"})
	// re- + définir needs hyphen form re-; "re-" prefix: pattern auto|auto-|re-|sur-
	// redéfinir: group re would need re- with hyphen. auto works without hyphen.
	// Use auto- with vowels:
	wt2 := tagging.MapWordTagger{
		"définir": {tagging.NewTaggedWord("définir", "V inf")},
	}
	got2 := NewFrenchTagger(wt2).Tag([]string{"auto-définir"})
	// auto- + définir: [^-].*... — définir starts with d not -, has vowels
	require.True(t, got2[0].IsTagged(), "auto-définir should tag via prefix")
	require.Equal(t, "V inf", *got2[0].GetReadings()[0].GetPOSTag())
	_ = got
}

func TestFrenchTagger_NounAdjPrefix(t *testing.T) {
	wt := tagging.MapWordTagger{
		"marché": {tagging.NewTaggedWord("marché", "N m s")},
	}
	got := NewFrenchTagger(wt).Tag([]string{"anti-marché"})
	require.True(t, got[0].IsTagged())
	require.Equal(t, "N m s", *got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "anti-marché", *got[0].GetReadings()[0].GetLemma())
}
