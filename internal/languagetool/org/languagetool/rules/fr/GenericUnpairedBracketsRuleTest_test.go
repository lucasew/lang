package fr

// Twin of languagetool-language-modules/fr/src/test/java/org/languagetool/rules/fr/GenericUnpairedBracketsRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGenericUnpairedBracketsRule_FrenchRule(t *testing.T) {
	rule := NewFrenchUnpairedBracketsRule(nil)
	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	require.Equal(t, 0, matchN("(Qu'est ce que c'est ?)"))
	// incorrect
	require.Equal(t, 1, matchN("(Qu'est ce que c'est ?"))
}
