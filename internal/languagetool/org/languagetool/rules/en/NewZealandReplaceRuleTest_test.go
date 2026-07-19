package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/NewZealandReplaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestNewZealandReplaceRule_Rule(t *testing.T) {
	rule := NewNewZealandReplaceRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Walk on the footpath."))))

	check := func(sentence, word string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(sentence))
		require.Equal(t, 1, len(matches), "sentence %q", sentence)
		require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
		require.Equal(t, word, matches[0].GetSuggestedReplacements()[0])
	}
	check("Sidewalk is not a place to park your car!", "Footpath")
	check("I walked on the sidewalk", "footpath")
}

// Java NewZealandReplaceRule: STYLE, LocaleViolation, sidewalk → footpath.
func TestNewZealandReplaceRule_Metadata(t *testing.T) {
	rule := NewNewZealandReplaceRule(nil)
	require.Equal(t, "EN_NZ_SIMPLE_REPLACE", rule.GetID())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSLocaleViolation, rule.GetLocQualityIssueType())
	require.Equal(t, []string{"footpath"}, rule.GetIncorrectExamples()[0].GetCorrections())
}
