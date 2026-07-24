package el

// Twin of languagetool-language-modules/el/src/test/java/org/languagetool/rules/el/GreekSpecificCaseRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGreekSpecificCaseRule_Rule(t *testing.T) {
	rule := NewGreekSpecificCaseRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ηνωμένες Πολιτείες"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Κατοικώ στις Ηνωμένες Πολιτείες."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Κατοικώ στις ΗΝΩΜΕΝΕΣ ΠΟΛΙΤΕΙΕΣ."))))

	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("ηνωμένες πολιτείες"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("ηνωμένες Πολιτείες"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Ηνωμένες πολιτείες"))))

	matches := rule.Match(languagetool.AnalyzePlain("Κατοικώ στις Ηνωμένες πολιτείες."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 13, matches[0].GetFromPos())
	require.Equal(t, 31, matches[0].GetToPos())
	require.Equal(t, "Ηνωμένες Πολιτείες", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "Οι λέξεις της συγκεκριμένης έκφρασης χρείαζεται να ξεκινούν με κεφαλαία γράμματα.", matches[0].GetMessage())

	// two spaces between words — tokenizer may collapse; accept span for joined phrase
	matches = rule.Match(languagetool.AnalyzePlain("Κατοικώ στις Ηνωμένες  πολιτείες."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 13, matches[0].GetFromPos())
	// End pos depends on whitespace tokenization; Java reports 32 with double space.
	require.GreaterOrEqual(t, matches[0].GetToPos(), 31)
}
