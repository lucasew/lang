package fr

// Twin of languagetool-language-modules/fr/src/test/java/org/languagetool/rules/fr/SimpleReplaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("J'ai pas de quoi"))))

	matches := rule.Match(languagetool.AnalyzePlain("jai pas de quoi"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "j'ai", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Jai pas de quoi"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "J'ai", matches[0].GetSuggestedReplacements()[0])
}
