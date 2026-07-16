package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianVerbConjugationRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of RussianVerbConjugationRuleTest.testRussianVerbConjugationRule
func TestRussianVerbConjugationRule_RussianVerbConjugationRule(t *testing.T) {
	r := NewRussianVerbConjugationRule(nil)
	require.Equal(t, "RU_VERB_CONJUGATION", r.GetID())
	ss := languagetool.SentenceStartTagName
	// Я + идёт (P3) → error
	p, vBad := "PNN:P1:Nom:Sin", "VB:Real:Imp:Tran:P3:Sin"
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("Я", &p, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("идёт", &vBad, nil), 2),
	})
	require.NotEmpty(t, r.Match(sent))
	// Я + иду (P1) → ok
	vOK := "VB:Real:Imp:Tran:P1:Sin"
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("Я", &p, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("иду", &vOK, nil), 2),
	})
	require.Empty(t, r.Match(sent2))
}
