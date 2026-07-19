package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/AmericanReplaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestAmericanReplaceRule_Rule(t *testing.T) {
	rule := NewAmericanReplaceRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Buy some gasoline."))))

	check := func(sentence, word string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(sentence))
		require.Equal(t, 1, len(matches), "sentence %q", sentence)
		require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
		require.Equal(t, word, matches[0].GetSuggestedReplacements()[0])
	}
	check("I love fish fingers.", "fish sticks")
}

// Java AmericanReplaceRule: STYLE, LocaleViolation, crisps → chips example.
func TestAmericanReplaceRule_Metadata(t *testing.T) {
	rule := NewAmericanReplaceRule(nil)
	require.Equal(t, "EN_US_SIMPLE_REPLACE", rule.GetID())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSLocaleViolation, rule.GetLocQualityIssueType())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Are baked <marker>crisps</marker> healthy?", inc[0].GetExample())
	require.Equal(t, []string{"chips"}, inc[0].GetCorrections())
}
