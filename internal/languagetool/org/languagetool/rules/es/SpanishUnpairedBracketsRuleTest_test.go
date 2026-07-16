package es

// Twin of languagetool-language-modules/es/src/test/java/org/languagetool/rules/es/SpanishUnpairedBracketsRuleTest.java
// Generic stack port; some apostrophe/name exceptions from Java are not modeled.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSpanishUnpairedBracketsRule_SpanishRule(t *testing.T) {
	rule := NewSpanishUnpairedBracketsRule(nil)
	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	require.Equal(t, 0, matchN("Soy un hombre (muy honrado)."))
	require.Equal(t, 1, matchN("Soy un hombre muy honrado)."))
	require.Equal(t, 1, matchN("Soy un hombre (muy honrado."))
	require.Equal(t, 1, matchN("Eso es (imposible. "))
	require.Equal(t, 1, matchN("Eso es) imposible. "))
}
