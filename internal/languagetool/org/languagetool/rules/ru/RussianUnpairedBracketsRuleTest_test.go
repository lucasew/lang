package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianUnpairedBracketsRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of RussianUnpairedBracketsRuleTest.testRuleRussian
func TestRussianUnpairedBracketsRule_RuleRussian(t *testing.T) {
	rule := NewRussianUnpairedBracketsRule(nil)
	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	// correct
	require.Equal(t, 0, matchN("(О жене и детях не беспокойся, я беру их на свои руки)."))
	require.Equal(t, 0, matchN("Позже выходит другая «южная поэма» «Бахчисарайский фонтан» (1824)."))
	require.Equal(t, 0, matchN("А \"б\" Д."))
	// incorrect: unpaired single quote
	require.Equal(t, 1, matchN("В таком ключе был начат в мае 1823 в Кишинёве роман в стихах 'Евгений Онегин."))
}
