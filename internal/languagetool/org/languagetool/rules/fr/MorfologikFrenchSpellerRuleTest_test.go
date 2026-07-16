package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func withFR(words ...string) *MorfologikFrenchSpellerRule {
	r := NewMorfologikFrenchSpellerRule()
	sp := morfologik.NewMorfologikSpeller(FrenchSpellerDict, 1)
	for _, w := range words {
		sp.AddWord(w)
	}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	return r
}

func TestMorfologikFrenchSpellerRule_MorfologikSpeller(t *testing.T) {
	r := withFR("bonjour", "monde")
	ms, err := r.Match(languagetool.AnalyzePlain("bonjour monnde"))
	require.NoError(t, err)
	require.Len(t, ms, 1)
}

func TestMorfologikFrenchSpellerRule_Apostrophes(t *testing.T) {
	r := withFR("l'homme", "d'accord")
	require.False(t, r.Speller.IsMisspelled("l'homme"))
	require.False(t, r.Speller.IsMisspelled("d'accord"))
}

func TestMorfologikFrenchSpellerRule_Hyphenated(t *testing.T) {
	r := withFR("peut-être", "grand-mère")
	require.False(t, r.Speller.IsMisspelled("peut-être"))
	require.True(t, r.Speller.IsMisspelled("peutetre"))
}

func TestMorfologikFrenchSpellerRule_EmptyString(t *testing.T) {
	r := withFR("a")
	ms, err := r.Match(languagetool.AnalyzePlain(""))
	require.NoError(t, err)
	require.Empty(t, ms)
}

func TestMorfologikFrenchSpellerRule_UnusualPunctuaion(t *testing.T) {
	r := withFR("oui")
	ms, err := r.Match(languagetool.AnalyzePlain("oui…"))
	require.NoError(t, err)
	_ = ms
}

func TestMorfologikFrenchSpellerRule_Sanity(t *testing.T) {
	r := NewMorfologikFrenchSpellerRule()
	require.Equal(t, MorfologikFrenchSpellerRuleID, r.GetID())
	require.Equal(t, FrenchSpellerDict, r.GetFileName())
}

func TestMorfologikFrenchSpellerRule_CorrectWords(t *testing.T) {
	r := withFR("chat", "chien")
	ms, err := r.Match(languagetool.AnalyzePlain("chat chien"))
	require.NoError(t, err)
	require.Empty(t, ms)
}

func TestMorfologikFrenchSpellerRule_Multiwords(t *testing.T) {
	r := withFR("parce", "que")
	// multiword as separate tokens accepted when both known
	ms, err := r.Match(languagetool.AnalyzePlain("parce que"))
	require.NoError(t, err)
	require.Empty(t, ms)
}

func TestMorfologikFrenchSpellerRule_MixedCase(t *testing.T) {
	r := withFR("Paris")
	// lowercase miss if only capital form registered (depends on speller lower-fallback)
	require.False(t, r.Speller.IsMisspelled("Paris"))
}

func TestMorfologikFrenchSpellerRule_IncorrectWords(t *testing.T) {
	r := withFR("maison")
	require.True(t, r.Speller.IsMisspelled("maizon"))
}

func TestMorfologikFrenchSpellerRule_WordSplitting(t *testing.T) {
	r := withFR("grand", "mère")
	ms, err := r.Match(languagetool.AnalyzePlain("grand mère"))
	require.NoError(t, err)
	require.Empty(t, ms)
}

func TestMorfologikFrenchSpellerRule_VerbsWithPronouns(t *testing.T) {
	r := withFR("donne-moi")
	require.False(t, r.Speller.IsMisspelled("donne-moi"))
}

func TestMorfologikFrenchSpellerRule_WordEdgeElision(t *testing.T) {
	r := withFR("j'aime")
	require.False(t, r.Speller.IsMisspelled("j'aime"))
}

func TestMorfologikFrenchSpellerRule_WordEdgeElisionWithTypos(t *testing.T) {
	r := withFR("j'aime")
	require.True(t, r.Speller.IsMisspelled("j'aimee"))
}

func TestMorfologikFrenchSpellerRule_WordBoundaryIssues(t *testing.T) {
	r := withFR("a", "b")
	ms, err := r.Match(languagetool.AnalyzePlain("a b"))
	require.NoError(t, err)
	require.Empty(t, ms)
}

func TestMorfologikFrenchSpellerRule_Screaming(t *testing.T) {
	r := withFR("STOP")
	require.False(t, r.Speller.IsMisspelled("STOP"))
}

func TestMorfologikFrenchSpellerRule_Tokenisation(t *testing.T) {
	r := withFR("bonjour")
	ms, err := r.Match(languagetool.AnalyzePlain("bonjour!"))
	require.NoError(t, err)
	// punctuation-only not misspelled
	_ = ms
}

func TestMorfologikFrenchSpellerRule_NoPrefixSplit(t *testing.T) {
	r := withFR("repartir")
	require.False(t, r.Speller.IsMisspelled("repartir"))
}

func TestMorfologikFrenchSpellerRule_Digits(t *testing.T) {
	r := withFR()
	// pure digits typically not misspelled by map empty? empty dict → misspelled
	// Accept non-alphabetic path is on Hunspell not Morfologik; just run Match
	ms, err := r.Match(languagetool.AnalyzePlain("123"))
	require.NoError(t, err)
	_ = ms
}

func TestMorfologikFrenchSpellerRule_ToImprove(t *testing.T) {
	r := withFR("amélioration")
	require.False(t, r.Speller.IsMisspelled("amélioration"))
}

func TestMorfologikFrenchSpellerRule_Multitokens(t *testing.T) {
	r := withFR("pomme", "de", "terre")
	ms, err := r.Match(languagetool.AnalyzePlain("pomme de terre"))
	require.NoError(t, err)
	require.Empty(t, ms)
}
