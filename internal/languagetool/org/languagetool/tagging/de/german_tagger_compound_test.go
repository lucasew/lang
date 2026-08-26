package de

import (
	"strings"
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

// Twin: VER:INF with first-upper rest-unchanged → SUB:…:INF; VER:INF only at index 0.
func TestPrefixPath_InfTitleNominalized(t *testing.T) {
	wt := tagging.MapWordTagger{
		"geben": {tagging.NewTaggedWord("geben", "VER:INF:NON")},
	}
	tagger := NewGermanTagger(wt)
	// mid-sentence Title infinitive
	got := tagger.Tag([]string{"Das", "Herumgeben"})
	// "Herumgeben" may hit verb expansion nominalized if attached — without expansion, prefix path
	tags := posTagsOf(got[1])
	// Without VerbExpansion, strip herum+geben via prefixesVerbs
	// need lastPart "geben" after strip
	require.Contains(t, tags, "SUB:NOM:SIN:NEU:INF")
	require.NotContains(t, tags, "VER:INF:NON", "mid-sentence Title INF must not keep VER:INF")
}

func TestPrefixPath_InfTitleAtStartKeepsVER(t *testing.T) {
	wt := tagging.MapWordTagger{
		"geben": {tagging.NewTaggedWord("geben", "VER:INF:NON")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"Herumgeben", "ist", "toll"})
	tags := posTagsOf(got[0])
	require.Contains(t, tags, "SUB:NOM:SIN:NEU:INF")
	require.Contains(t, tags, "VER:INF:NON")
}

func TestIsFirstCharUpperRestUnchanged(t *testing.T) {
	require.True(t, isFirstCharUpperRestUnchanged("Herumgeben"))
	require.True(t, isFirstCharUpperRestUnchanged("HerumGeben")) // rest unchanged may be mixed
	require.False(t, isFirstCharUpperRestUnchanged("herumgeben"))
	// Java: ALLCAPS also matches (first.toUpper + rest == word)
	require.True(t, isFirstCharUpperRestUnchanged("HERUMGEBEN"))
	require.True(t, isFirstCharUpperRestUnchanged("Übergeben"))
	require.False(t, isFirstCharUpperRestUnchanged("üBergeben"))
}

// Twin: EIZ lemma uses firstPart + base lemma without firstPart.toLowerCase (Java).
func TestPrefixPath_EIZ_LemmaKeepsFirstPartCase(t *testing.T) {
	wt := tagging.MapWordTagger{
		"geben": {tagging.NewTaggedWord("geben", "VER:INF:NON")},
	}
	tagger := NewGermanTagger(wt)
	// Title "Einzugeben" is title-or-lower; lastPart zugeben→zu+geben
	got := tagger.Tag([]string{"Einzugeben"})
	// firstPart = removeEnd("Einzugeben","zugeben") = "Ein"
	found := false
	for _, r := range got[0].GetReadings() {
		if r.GetPOSTag() != nil && strings.Contains(*r.GetPOSTag(), "EIZ") {
			require.NotNil(t, r.GetLemma())
			require.Equal(t, "Eingeben", *r.GetLemma())
			found = true
		}
	}
	require.True(t, found, "expected EIZ reading, tags=%v", posTagsOf(got[0]))
}

// Twin: firstPart.equals("un") case-sensitive — "Un…" is not skipped.
func TestPrefixPath_UnCaseSensitive(t *testing.T) {
	wt := tagging.MapWordTagger{
		"beeindruckt": {tagging.NewTaggedWord("beeindrucken", "VER:PA2:SFT")},
	}
	tagger := NewGermanTagger(wt)
	// lower un* skips VER
	got := tagger.Tag([]string{"unbeeindruckt"})
	for _, tg := range posTagsOf(got[0]) {
		require.False(t, strings.HasPrefix(tg, "VER") && !strings.HasPrefix(tg, "VER:PA"), "lower un should skip bare VER")
	}
}

// Twin: durch dual PA2 requires firstPart.equals("durch") exact (not "Durch").
func TestDurch_TitleFirstPartNoDualPA2(t *testing.T) {
	wt := tagging.MapWordTagger{
		"läuft": {tagging.NewTaggedWord("laufen", "VER:3:SIN:PRÄ:SFT")},
	}
	tagger := NewGermanTagger(wt)
	// "Durchläuft" is title-or-lower; firstPart removeEnd = "Durch" ≠ "durch"
	got := tagger.Tag([]string{"Durchläuft"})
	tags := posTagsOf(got[0])
	require.NotContains(t, tags, "VER:PA2:SFT")
	require.NotContains(t, tags, "PA2:PRD:GRU:VER")
}

// Twin: dash-linked upper surface uses addStem + simple getAnalyzedTokens (lemma stem+lemma).
func TestDashLinked_UpperStemLemma(t *testing.T) {
	wt := tagging.MapWordTagger{
		"Zentrum": {tagging.NewTaggedWord("Zentrum", "SUB:NOM:SIN:NEU")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"Diabetes-Zentrum"})
	require.Contains(t, posTagsOf(got[0]), "SUB:NOM:SIN:NEU")
	// stem "Diabetes-" ends with '-' → addStem does not lowercase SUB lemma
	var lemma string
	for _, r := range got[0].GetReadings() {
		if r.GetLemma() != nil {
			lemma = *r.GetLemma()
		}
	}
	require.Equal(t, "Diabetes-Zentrum", lemma)
}

// Twin: javaRemoveEnd case-sensitive suffix strip.
func TestJavaRemoveEnd(t *testing.T) {
	require.Equal(t, "Ein", javaRemoveEnd("Einlädst", "lädst"))
	require.Equal(t, "Einlädst", javaRemoveEnd("Einlädst", "Lädst")) // case-sensitive miss
	require.Equal(t, "durch", javaRemoveEnd("durchläuft", "läuft"))
}
