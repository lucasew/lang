package be

// Twin of languagetool-language-modules/be/src/test/java/org/languagetool/rules/be/SimpleReplaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Напрамкі дзейнасці былі ўзгоднены з ,"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Яго камп’ютар выключыўся."))))

	matches := rule.Match(languagetool.AnalyzePlain("Яго кампутар выключыўся."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
	require.Equal(t, "камп’ютар", matches[0].GetSuggestedReplacements()[0])
}
