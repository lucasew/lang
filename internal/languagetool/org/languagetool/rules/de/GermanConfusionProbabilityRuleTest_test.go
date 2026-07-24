package de

// Twin of GermanConfusionProbabilityRuleTest.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	"github.com/stretchr/testify/require"
)

func TestGermanConfusionProbabilityRule_Constructor(t *testing.T) {
	r := NewGermanConfusionProbabilityRule(nil)
	require.NotNil(t, r)
	require.Equal(t, "DE_CONFUSION_RULE", r.GetID())
	// Without LM, Match must not invent hits.
	require.Empty(t, r.Match(languagetool.AnalyzePlain("Während Sie das Ganze mir einem Holzlöffel rühren.")))
}

func TestGermanConfusionProbabilityRule_WithLMAndPair(t *testing.T) {
	// Fake LM: always prefer "mit" over "mir" in 3-grams (higher prob for contexts with mit).
	lm := ngrams.FuncLanguageModel(func(tokens []string) ngrams.Probability {
		joined := ""
		for _, t := range tokens {
			joined += t + " "
		}
		if containsToken(tokens, "mit") {
			return ngrams.NewProbabilitySimple(0.9, 1.0)
		}
		if containsToken(tokens, "mir") {
			return ngrams.NewProbabilitySimple(0.001, 1.0)
		}
		return ngrams.NewProbabilitySimple(0.1, 1.0)
	})
	r := NewGermanConfusionProbabilityRuleWithLM(lm)
	// Force pair mir/mit for the test (like Java setConfusionPair).
	pair := rules.NewConfusionPairTokens("mir", "mit", 10, true)
	r.SetConfusionPair(pair)
	matches := r.Match(languagetool.AnalyzePlain("Während Sie das Ganze mir einem Holzlöffel rühren."))
	require.NotEmpty(t, matches)
	require.Contains(t, matches[0].GetSuggestedReplacements(), "mit")
}

func containsToken(tokens []string, want string) bool {
	for _, t := range tokens {
		if t == want {
			return true
		}
	}
	return false
}

func TestGermanConfusionProbabilityRule_SentenceException(t *testing.T) {
	lm := ngrams.UniformLanguageModel(0.5, 1)
	r := NewGermanConfusionProbabilityRuleWithLM(lm)
	// Sentence exception pattern "wir (" should skip via IsException.
	require.True(t, r.IsException("Hallo, wir (die Dingsbums Gmbh)", 0, 3))
}

func TestGermanConfusionAntiPatternsCount(t *testing.T) {
	// Java GermanConfusionProbabilityRule.ANTI_PATTERNS length
	require.Len(t, GermanConfusionAntiPatterns, 11)
}

func TestGermanConfusionAntiPatternImmunizeFasstZusammen(t *testing.T) {
	// token-only anti-pattern: fasst … zusammen
	toks := withPositions(
		sentStartATR(),
		atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "der"),
		atrWithPOS("Übersicht", "SUB:NOM:SIN:FEM", "Übersicht"),
		atrWithPOS("fasst", "VER:3:SIN:PRS:NON", "fassen"),
		atrWithPOS("Ziele", "SUB:AKK:PLU:NEU", "Ziel"),
		atrWithPOS("zusammen", "ZUS", "zusammen"),
		atrWithPOS(".", "PKT", "."),
	)
	sent := languagetool.NewAnalyzedSentence(toks)
	fasst := toks[3]
	require.True(t, deConfusionIsCoveredByAntiPattern(sent, fasst.GetStartPos(), fasst.GetEndPos()))
}
