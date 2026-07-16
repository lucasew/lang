package pl

// Twin of languagetool-language-modules/pl/src/test/java/org/languagetool/rules/pl/PolishUnpairedBracketsRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPolishUnpairedBracketsRule_RulePolish(t *testing.T) {
	rule := NewPolishUnpairedBracketsRule(nil)
	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	require.Equal(t, 0, matchN("(To jest zdanie do testowania)."))
	require.Equal(t, 0, matchN("A \"B\" C."))
	require.Equal(t, 0, matchN("\"A\" B \"C\"."))
	require.Equal(t, 1, matchN("W tym zdaniu jest niesparowany „cudzysłów."))
}
