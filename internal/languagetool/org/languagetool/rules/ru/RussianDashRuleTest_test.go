package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianDashRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRussianDashRule_Rule(t *testing.T) {
	rule := NewRussianDashRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Он вышел из-за забора."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ростов-на-Дону."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("ведром — работай"))))

	matches := rule.Match(languagetool.AnalyzePlain("из—за"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "из-за", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Ростов — на — Дону"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Ростов-на-Дону", matches[0].GetSuggestedReplacements()[0])
}
