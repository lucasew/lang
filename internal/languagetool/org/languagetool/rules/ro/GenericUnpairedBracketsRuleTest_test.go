package ro

// Twin of languagetool-language-modules/ro/src/test/java/org/languagetool/rules/ro/GenericUnpairedBracketsRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGenericUnpairedBracketsRule_RomanianRule(t *testing.T) {
	rule := NewRomanianUnpairedBracketsRule(nil)
	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	require.Equal(t, 0, matchN("A fost plecat (pentru puțin timp)."))
	require.Equal(t, 0, matchN("A fost plecat pentru „puțin timp”."))
	// incorrect
	require.Equal(t, 1, matchN("A fost plecat „pentru... puțin timp."))
	require.Equal(t, 1, matchN("A fost plecat «puțin."))
}
