package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestWordCoherencyRule(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	// Java example: pesebre … pessebre (surface default replacement)
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Un pesebre ací i un altre pessebre allà."),
	})
	require.NotEmpty(t, matches)
	require.Equal(t, "pesebre", matches[0].GetSuggestedReplacements()[0])
}

func TestWordCoherencyRule_SynthesizeInflection(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	// When second form has VAND POS, synthesize other spelling with that tag.
	rule.Synthesize = func(lemma, postag string) []string {
		if lemma == "pesebre" && postag == "NCMS000" {
			return []string{"pesebre-synth"}
		}
		return nil
	}
	// Second occurrence needs POS so createReplacement uses synth
	sent := languagetool.AnalyzeWithTagger(
		"Un pesebre ací i un altre pessebre allà.",
		func(tok string) []languagetool.TokenTag {
			if tok == "pessebre" {
				return []languagetool.TokenTag{{POS: "NCMS000", Lemma: "pessebre"}}
			}
			if tok == "pesebre" {
				return []languagetool.TokenTag{{POS: "NCMS000", Lemma: "pesebre"}}
			}
			return nil
		},
	)
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{sent})
	require.NotEmpty(t, matches)
	require.Equal(t, "pesebre-synth", matches[0].GetSuggestedReplacements()[0])
}

func TestWordCoherencyRule_SynthFailFallsBack(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	rule.Synthesize = func(lemma, postag string) []string { return nil }
	sent := languagetool.AnalyzeWithTagger(
		"Un pesebre ací i un altre pessebre allà.",
		func(tok string) []languagetool.TokenTag {
			if tok == "pessebre" || tok == "pesebre" {
				return []languagetool.TokenTag{{POS: "NCMS000", Lemma: tok}}
			}
			return nil
		},
	)
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{sent})
	require.NotEmpty(t, matches)
	require.Equal(t, "pesebre", matches[0].GetSuggestedReplacements()[0])
}

func TestWordCoherencyValencianRule(t *testing.T) {
	rule := NewWordCoherencyValencianRule(nil)
	// Java example: Este … aquest
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Este home d'ací parla amb aquest altre ací."),
	})
	require.NotEmpty(t, matches)
}

func TestWordCoherencyValencianRule_Synthesize(t *testing.T) {
	rule := NewWordCoherencyValencianRule(nil)
	rule.Synthesize = func(lemma, postag string) []string {
		if lemma == "este" && postag == "DD0MS0" {
			return []string{"este-synth"}
		}
		return nil
	}
	sent := languagetool.AnalyzeWithTagger(
		"Este home d'ací parla amb aquest altre ací.",
		func(tok string) []languagetool.TokenTag {
			switch tok {
			case "Este":
				return []languagetool.TokenTag{{POS: "DD0MS0", Lemma: "este"}}
			case "aquest":
				return []languagetool.TokenTag{{POS: "DD0MS0", Lemma: "aquest"}}
			default:
				return nil
			}
		},
	)
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{sent})
	require.NotEmpty(t, matches)
	// marked surface is "aquest" (lowercase) → no UppercaseFirstChar
	require.Equal(t, "este-synth", matches[0].GetSuggestedReplacements()[0])
}
