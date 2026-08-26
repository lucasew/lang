package ru

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Unit helpers: exercise conjugation helpers + Match control flow with fixed readings
// (same POS shape as real russian.dict first readings: PNN:Sin:Nom:P1, VB:…:Sin:P3).

func TestRussianVerbConjugationRule_Mismatch(t *testing.T) {
	r := NewRussianVerbConjugationRule(nil)
	require.Equal(t, "RU_VERB_CONJUGATION", r.GetID())

	// Я (Sin:Nom:P1) + идёт (Sin:P3) → mismatch person (right slot)
	ss := languagetool.SentenceStartTagName
	p := "PNN:Sin:Nom:P1"
	v := "VB:Real:INTR:IMPFV:Sin:P3"
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("Я", &p, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("идёт", &v, nil), 2),
	})
	matches := r.Match(sent)
	require.Len(t, matches, 1)
	require.Equal(t, "Неверное спряжение глагола или неверное местоимение", matches[0].GetMessage())
	require.Equal(t, "Неверное спряжение глагола", matches[0].GetShortMessage())
}

func TestRussianVerbConjugationRule_OK(t *testing.T) {
	r := NewRussianVerbConjugationRule(nil)
	ss := languagetool.SentenceStartTagName
	p := "PNN:Sin:Nom:P1"
	v := "VB:Real:INTR:IMPFV:Sin:P1"
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("Я", &p, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("иду", &v, nil), 2),
	})
	require.Empty(t, r.Match(sent))
}

func TestRussianVerbConjugationRule_SkipIConjunction(t *testing.T) {
	// "… и я согласились" — previous token "и" skips pronoun/verb pair.
	r := NewRussianVerbConjugationRule(nil)
	ss := languagetool.SentenceStartTagName
	conj := "CONJ"
	p := "PNN:Sin:Nom:P1"
	// deliberately wrong past PL after "я" — should still skip due to "и"
	v := "VB:Past:INTR:PFV:PL"
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("и", &conj, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("я", &p, nil), 2),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("согласились", &v, nil), 4),
	})
	require.Empty(t, r.Match(sent))
}

func TestRussianVerbConjugationRule_SkipMozhetByt(t *testing.T) {
	// "Ты может быть …" — next2==быть && next==может skips.
	r := NewRussianVerbConjugationRule(nil)
	ss := languagetool.SentenceStartTagName
	p := "PNN:Sin:Nom:P2"
	v := "VB:Real:INTR:IMPFV:Sin:P3" // person mismatch would otherwise fire
	inf := "VB:INF:INTR:IMPFV"
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("Ты", &p, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("может", &v, nil), 3),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("быть", &inf, nil), 9),
	})
	require.Empty(t, r.Match(sent))
}

func TestConjugationHelpers(t *testing.T) {
	// Present/future: right slot person; left is Sin/PL (real RU tag order).
	require.True(t, isConjugationInPresentOrFutureWrong("Sin", "P1", "Sin", "P3"))
	require.False(t, isConjugationInPresentOrFutureWrong("Sin", "P1", "Sin", "P1"))
	// Gendered 3rd person: only wrong when verb number is PL.
	require.True(t, isConjugationInPresentOrFutureWrong("Fem", "P3", "PL", "P3"))
	require.False(t, isConjugationInPresentOrFutureWrong("Fem", "P3", "Sin", "P3"))
	// Past: Sin pronoun rejects PL/Neut only.
	require.True(t, isConjugationInPastWrong("Sin", "PL"))
	require.True(t, isConjugationInPastWrong("Sin", "Neut"))
	require.False(t, isConjugationInPastWrong("Sin", "Masc"))
	require.False(t, isConjugationInPastWrong("Sin", "Fem"))
	require.True(t, isConjugationInPastWrong("PL", "Masc"))
	require.False(t, isConjugationInPastWrong("PL", "PL"))
	require.True(t, isConjugationInPastWrong("Fem", "Masc"))
	require.False(t, isConjugationInPastWrong("Fem", "Fem"))
}
