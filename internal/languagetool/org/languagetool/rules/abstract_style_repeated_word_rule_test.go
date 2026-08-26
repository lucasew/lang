package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAbstractStyleRepeatedWordRule(t *testing.T) {
	r := NewAbstractStyleRepeatedWordRule()
	// Java matches via lemmas (or isPartOfWord); inject lemmas, no surface invent
	s1 := languagetool.AnalyzeWithTagger("The cat sat near the cat.", func(tok string) []languagetool.TokenTag {
		if tok == "cat" {
			return []languagetool.TokenTag{{POS: "NN", Lemma: "cat"}}
		}
		return nil
	})
	matches := r.MatchList([]*languagetool.AnalyzedSentence{s1})
	require.NotEmpty(t, matches)

	s2a := languagetool.AnalyzeWithTagger("Dogs are nice.", func(tok string) []languagetool.TokenTag {
		if tok == "Dogs" {
			return []languagetool.TokenTag{{POS: "NNS", Lemma: "dog"}}
		}
		return nil
	})
	s2b := languagetool.AnalyzeWithTagger("Dogs run fast.", func(tok string) []languagetool.TokenTag {
		if tok == "Dogs" {
			return []languagetool.TokenTag{{POS: "NNS", Lemma: "dog"}}
		}
		return nil
	})
	m2 := r.MatchList([]*languagetool.AnalyzedSentence{s2a, s2b})
	require.NotEmpty(t, m2)

	// without lemmas: fail closed (no surface EqualFold invent)
	require.Empty(t, r.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("The cat sat near the cat."),
	}))
}
