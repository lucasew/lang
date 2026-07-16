package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianSimpleReplaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRussianSimpleReplaceRule_Rule(t *testing.T) {
	rule := NewRussianSimpleReplaceRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Рост кораллов тут самый быстрый,"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Книга была порвана."))))

	matches := rule.Match(languagetool.AnalyzePlain("Книга была порвата."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
	require.Equal(t, "порвана", matches[0].GetSuggestedReplacements()[0])
}
