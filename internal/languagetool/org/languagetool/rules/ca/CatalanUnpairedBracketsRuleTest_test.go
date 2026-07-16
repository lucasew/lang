package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/rules/ca/CatalanUnpairedBracketsRuleTest.java
// Reduced surface cases (full Java has many exceptions).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCatalanUnpairedBracketsRule_Rule(t *testing.T) {
	rule := NewCatalanUnpairedBracketsRule(nil)
	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	require.Equal(t, 0, matchN("Això és (correcte)."))
	require.Equal(t, 1, matchN("Això és (incorrecte."))
	require.Equal(t, 1, matchN("Això és incorrecte)."))
}
