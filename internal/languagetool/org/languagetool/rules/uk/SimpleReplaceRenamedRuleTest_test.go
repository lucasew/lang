package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/SimpleReplaceRenamedRuleTest.java
// Surface + light inflection matching (no POS/lemma tagger).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceRenamedRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceRenamedRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Київ."))))

	matches := rule.Match(languagetool.AnalyzePlain("Дніпродзержинська"))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetSuggestedReplacements(), "Кам'янське")
	require.Contains(t, matches[0].GetSuggestedReplacements(), "Кам'янський")
	require.Contains(t, matches[0].GetMessage(), "2016")

	matches = rule.Match(languagetool.AnalyzePlain("дніпродзержинського."))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetSuggestedReplacements(), "кам'янський")

	matches = rule.Match(languagetool.AnalyzePlain("Червонознам'янка."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Знам'янка", "Знаменка"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(languagetool.AnalyzePlain("Переяслав-Хмельницький."))
	require.Equal(t, 1, len(matches))
	// Case-variant keys merge noun + adj suggestions (Java uses separate lemma readings).
	require.Contains(t, matches[0].GetSuggestedReplacements(), "Переяслав")
	require.Contains(t, matches[0].GetSuggestedReplacements(), "Переяславський")
}
