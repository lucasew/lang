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
func TestCompoundSplit_NonPrefixDropsVER(t *testing.T) {
	wt := tagging.MapWordTagger{
		"Haus": {tagging.NewTaggedWord("Haus", "SUB:NOM:SIN:NEU")},
		// also a VER reading on last part would be dropped when first is not prfx
		"bau": {tagging.NewTaggedWord("bauen", "VER:INF:NON"), tagging.NewTaggedWord("Bau", "SUB:NOM:SIN:MAS")},
	}
	tagger := NewGermanTagger(wt)
	tagger.SplitCompound = func(word string) []string {
		if word == "Hochbau" {
			return []string{"Hoch", "bau"}
		}
		return nil
	}
	// Uppercase word → last part uppercased to Bau; TagWordExact("Bau") may miss — tag "bau" lower via no upper on non-start?
	// Java: startsWithUppercase(word) → uppercaseFirstChar(lastPart) → "Bau"
	// Our map has "bau" only for last after upper → "Bau" miss.
	// Use all-lower surface so last stays "bau".
	tagger2 := NewGermanTagger(wt)
	tagger2.SplitCompound = func(word string) []string {
		if word == "hochbau" {
			return []string{"hoch", "bau"}
		}
		return nil
	}
	got := tagger2.Tag([]string{"hochbau"})
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
