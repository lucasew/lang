package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianSpecificCaseRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRussianSpecificCaseRule_Rule(t *testing.T) {
	rule := NewRussianSpecificCaseRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Рытый Банк"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Центральный банк РФ"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Рытый банк"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("центральный банк РФ"))))

	matches := rule.Match(languagetool.AnalyzePlain("I like air France."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 7, matches[0].GetFromPos())
	require.Equal(t, 17, matches[0].GetToPos())
	require.Equal(t, "Air France", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "Для специальных наименований используйте начальную заглавную букву.", matches[0].GetMessage())
}
