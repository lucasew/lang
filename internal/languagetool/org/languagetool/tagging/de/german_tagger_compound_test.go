package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestSanitizeWord_DashCompound(t *testing.T) {
	wt := tagging.MapWordTagger{
		"Zentrum": {tagging.NewTaggedWord("Zentrum", "SUB:NOM:SIN:NEU")},
	}
	tagger := NewGermanTagger(wt)
	// last part noun → sanitize to Zentrum
	require.Equal(t, "Zentrum", tagger.sanitizeWord("Diabetes-Zentrum"))
	// ending dash unchanged
	require.Equal(t, "foo-", tagger.sanitizeWord("foo-"))
}

func TestAddStem(t *testing.T) {
	in := []tagging.TaggedWord{tagging.NewTaggedWord("Zentrum", "SUB:NOM:SIN:NEU")}
	got := addStem(in, "Diabetes-")
	// SUB lemma lowercased when stem does not end with '-' — stem ends with '-' so no lower
	require.Equal(t, "Diabetes-Zentrum", got[0].Lemma)
	got2 := addStem(in, "Diabetes")
	require.Equal(t, "Diabeteszentrum", got2[0].Lemma) // lower lemma
}

func TestDashLinkedTagging(t *testing.T) {
	wt := tagging.MapWordTagger{
		"Zentrum": {tagging.NewTaggedWord("Zentrum", "SUB:NOM:SIN:NEU")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"Diabetes-Zentrum"})
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "SUB:NOM:SIN:NEU", *got[0].GetReadings()[0].GetPOSTag())
	// lemma includes stem
	require.NotNil(t, got[0].GetReadings()[0].GetLemma())
	require.Contains(t, *got[0].GetReadings()[0].GetLemma(), "Zentrum")
}

func TestSeparablePrefix_NEB(t *testing.T) {
	// einlädst style: prefix ein + lädst
	wt := tagging.MapWordTagger{
		"lädst": {tagging.NewTaggedWord("laden", "VER:2:SIN:PRÄ:NON")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"einlädst"})
	tags := posTagsOf(got[0])
	require.NotEmpty(t, tags)
	found := false
	for _, tg := range tags {
		if stringsHasSuffix(tg, ":NEB") && stringsHasPrefix(tg, "VER:2") {
			found = true
		}
	}
	require.True(t, found, "expected VER:2…:NEB, got %v", tags)
}

func TestSeparablePrefix_ZuEIZ(t *testing.T) {
	wt := tagging.MapWordTagger{
		"geben": {tagging.NewTaggedWord("geben", "VER:INF:NON")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"einzugeben"})
	tags := posTagsOf(got[0])
	found := false
	for _, tg := range tags {
		if stringsContains(tg, "EIZ") {
			found = true
		}
	}
	require.True(t, found, "expected EIZ, got %v", tags)
}

func TestMatchesUppercaseAdjective(t *testing.T) {
	wt := tagging.MapWordTagger{
		"schön": {tagging.NewTaggedWord("schön", "ADJ:PRD:GRU")},
	}
	tagger := NewGermanTagger(wt)
	require.True(t, tagger.matchesUppercaseAdjective("Schön"))
	require.False(t, tagger.matchesUppercaseAdjective("Tisch"))
}

func stringsHasSuffix(s, p string) bool {
	return len(s) >= len(p) && s[len(s)-len(p):] == p
}
func stringsContains(s, p string) bool {
	return len(s) >= len(p) && (s == p || len(p) == 0 ||
		func() bool {
			for i := 0; i+len(p) <= len(s); i++ {
				if s[i:i+len(p)] == p {
					return true
				}
			}
			return false
		}())
}
