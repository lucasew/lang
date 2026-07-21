package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

// Twin compoundParts + prfxs.contains: separable first part → :NEB + lemma first+baseLemma.
func TestCompoundSplit_PrefixVerbNEB(t *testing.T) {
	wt := tagging.MapWordTagger{
		"lädst": {tagging.NewTaggedWord("laden", "VER:2:SIN:PRÄ:NON")},
	}
	tagger := NewGermanTagger(wt)
	tagger.SplitCompound = func(word string) []string {
		if word == "einlädst" {
			return []string{"ein", "lädst"}
		}
		return nil
	}
	got := tagger.Tag([]string{"einlädst"})
	tags := posTagsOf(got[0])
	require.Contains(t, tags, "VER:2:SIN:PRÄ:NON:NEB")
	// lemma = firstPart + tag.lemma
	var lemma string
	for _, r := range got[0].GetReadings() {
		if r.GetPOSTag() != nil && *r.GetPOSTag() == "VER:2:SIN:PRÄ:NON:NEB" && r.GetLemma() != nil {
			lemma = *r.GetLemma()
		}
	}
	require.Equal(t, "einladen", lemma)
}

// Non-prefix compound: VER tags filtered out (Java stream filter !contains VER).
// firstPart must not be in prfxs ("Auto" is not a separable verb prefix).
func TestCompoundSplit_NonPrefixDropsVER(t *testing.T) {
	wt := tagging.MapWordTagger{
		"bau": {
			tagging.NewTaggedWord("bauen", "VER:INF:NON"),
			tagging.NewTaggedWord("Bau", "SUB:NOM:SIN:MAS"),
		},
	}
	tagger := NewGermanTagger(wt)
	tagger.SplitCompound = func(word string) []string {
		if word == "autobau" {
			return []string{"auto", "bau"}
		}
		return nil
	}
	got := tagger.Tag([]string{"autobau"})
	tags := posTagsOf(got[0])
	for _, tg := range tags {
		require.NotContains(t, tg, "VER", "non-prefix compound must drop VER tags")
	}
	require.Contains(t, tags, "SUB:NOM:SIN:MAS")
}

// Mid-sentence title (not first-char-lower): no :NEB on prefix compound finite.
func TestCompoundSplit_TitleMidNoNEB(t *testing.T) {
	wt := tagging.MapWordTagger{
		"lädst": {tagging.NewTaggedWord("laden", "VER:2:SIN:PRÄ:NON")},
	}
	tagger := NewGermanTagger(wt)
	tagger.SplitCompound = func(word string) []string {
		if word == "Einlädst" {
			return []string{"ein", "lädst"}
		}
		return nil
	}
	got := tagger.Tag([]string{"Dann", "Einlädst"})
	tags := posTagsOf(got[1])
	// firstPart "ein" is prfx; Title surface fails first-char-lower and not index 0
	// → else if !IMP adds plain without NEB
	require.NotContains(t, tags, "VER:2:SIN:PRÄ:NON:NEB")
	require.Contains(t, tags, "VER:2:SIN:PRÄ:NON")
}
