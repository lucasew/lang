package el

// Twin of languagetool-language-modules/el/src/test/java/org/languagetool/rules/el/GreekRedundancyRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGreekRedundancyRule_Rule(t *testing.T) {
	rule := NewGreekRedundancyRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Τώρα μπαίνω στο σπίτι."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Απόψε θα βγω."))))
}

func TestGreekRedundancyRule_RuleWithinSentence(t *testing.T) {
	rule := NewGreekRedundancyRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Τώρα μπαίνω μέσα στο σπίτι."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "μπαίνω", matches[0].GetSuggestedReplacements()[0])
}

func TestGreekRedundancyRule_RuleBegginingOfSentence(t *testing.T) {
	rule := NewGreekRedundancyRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Απόψε το βράδυ θα βγω."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Απόψε", matches[0].GetSuggestedReplacements()[0])
}

func TestGreekRedundancyRule_RuleMultipleSuggestions(t *testing.T) {
	rule := NewGreekRedundancyRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Το μαγαζί ήταν ωραίο, αλλά όμως δεν πέρασα καλά."))
	require.Equal(t, 1, len(matches))
	// File stores a single replacement string containing a comma (not pipe-separated).
	require.Equal(t, "αλλά,όμως", matches[0].GetSuggestedReplacements()[0])
}
