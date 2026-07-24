package es

// Twin of languagetool-language-modules/es/src/test/java/org/languagetool/rules/es/SimpleReplaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("sanitización"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "desinfección", matches[0].GetSuggestedReplacements()[0])

	matches2 := rule.Match(languagetool.AnalyzePlain("Esta frase no tiene errores."))
	require.Equal(t, 0, len(matches2))
}
