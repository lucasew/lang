package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/SimpleReplaceSoftRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceSoftRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceSoftRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ці рядки повинні збігатися."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("у Трускавці."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("завидна"))))

	matches := rule.Match(languagetool.AnalyzePlain("Цей брелок"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"дармовис"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(languagetool.AnalyzePlain("Не знайде спасіння."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"рятування", "рятунок", "порятунок", "визволення"}, matches[0].GetSuggestedReplacements())
	require.Contains(t, matches[0].GetMessage(), "релігія")
}

func TestSimpleReplaceSoftRule_RuleForDerivats(t *testing.T) {
	rule := NewSimpleReplaceSoftRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Підключивши"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Увімкнувши", "Під'єднавши", "Приєднавши"}, matches[0].GetSuggestedReplacements())
}
