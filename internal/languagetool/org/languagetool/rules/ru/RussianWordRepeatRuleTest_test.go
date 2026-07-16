package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianWordRepeatRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRussianWordRepeatRule_Rule(t *testing.T) {
	rule := NewRussianWordRepeatRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Повтор слов в предложении."))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Повтор слов в повтор предложении."))))
}
