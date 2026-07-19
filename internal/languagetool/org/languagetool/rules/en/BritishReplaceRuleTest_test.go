package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/BritishReplaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestBritishReplaceRule_Rule(t *testing.T) {
	rule := NewBritishReplaceRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Buy some petrol."))))

	check := func(sentence, word string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(sentence))
		require.Equal(t, 1, len(matches), "sentence %q", sentence)
		require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
		require.Equal(t, word, matches[0].GetSuggestedReplacements()[0])
	}
	check("Diapers for sale!", "Nappies")
	check("We have some diapers for sale.", "nappies")
	check("Everyone loves cotton candy.", "candy floss")
}

// Java BritishReplaceRule: STYLE, LocaleViolation, drapes → curtains example.
func TestBritishReplaceRule_Metadata(t *testing.T) {
	rule := NewBritishReplaceRule(nil)
	require.Equal(t, "EN_GB_SIMPLE_REPLACE", rule.GetID())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSLocaleViolation, rule.GetLocQualityIssueType())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "We can produce <marker>drapes</marker> of any size or shape from a choice of over 500 different fabrics.", inc[0].GetExample())
	require.Equal(t, []string{"curtains"}, inc[0].GetCorrections())
}
