package de

// Twin of GermanStyleRepeatedWordRuleTest (surface form matching).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanStyleRepeatedWordRule_Rule(t *testing.T) {
	rule := NewGermanStyleRepeatedWordRule(nil)
	matchN := func(text string) int {
		sents := languagetool.SplitAndAnalyze(text)
		return len(rule.MatchList(sents))
	}
	// "großen" in both sentences → 2 matches
	require.Equal(t, 2, matchN("Der alte Mann wohnte in einem großen Haus. Es stand in einem großen Garten."))
	// different adjectives
	require.Equal(t, 0, matchN("Der alte Mann wohnte in einem großen Haus. Es stand in einem weitläufigen Garten."))
	// soft: other Java goods often rely on lemma/POS; surface may differ slightly
	require.Equal(t, 0, matchN("Endlos lang zog sich der Ton dahin, aber schließlich verklang er doch."))
}
