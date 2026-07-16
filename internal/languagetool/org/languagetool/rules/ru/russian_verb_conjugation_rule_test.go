package ru

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRussianVerbConjugationRule_Mismatch(t *testing.T) {
	r := NewRussianVerbConjugationRule(nil)
	require.Equal(t, "RU_VERB_CONJUGATION", r.GetID())

	// Я (P1:Sin) + идёт (P3:Sin) → mismatch person
	ss := languagetool.SentenceStartTagName
	p := "PNN:P1:Nom:Sin"
	v := "VB:Real:Imp:Tran:P3:Sin"
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("Я", &p, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("идёт", &v, nil), 2),
	})
	matches := r.Match(sent)
	require.Len(t, matches, 1)
}

func TestRussianVerbConjugationRule_OK(t *testing.T) {
	r := NewRussianVerbConjugationRule(nil)
	ss := languagetool.SentenceStartTagName
	p := "PNN:P1:Nom:Sin"
	v := "VB:Real:Imp:Tran:P1:Sin"
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("Я", &p, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("иду", &v, nil), 2),
	})
	require.Empty(t, r.Match(sent))
}

func TestConjugationHelpers(t *testing.T) {
	require.True(t, isConjugationPresentFutureWrong("P1", "Sin", "P3", "Sin"))
	require.False(t, isConjugationPresentFutureWrong("P1", "Sin", "P1", "Sin"))
	require.True(t, isConjugationPastWrong("Sin", "PL"))
	require.False(t, isConjugationPastWrong("PL", "PL"))
}
