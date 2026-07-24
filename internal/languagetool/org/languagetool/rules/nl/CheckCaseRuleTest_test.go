package nl

// Twin of languagetool-language-modules/nl/src/test/java/org/languagetool/rules/nl/CheckCaseRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCheckCaseRule_Rule(t *testing.T) {
	rule := NewCheckCaseRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("een bisschop"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Een bisschop"))))
	// ignored all-upper long phrase
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("EEN BISSCHOP"))))

	matches := rule.Match(languagetool.AnalyzePlain("Een Bisschop"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Een bisschop", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Hij is een Bisschop."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "een bisschop", matches[0].GetSuggestedReplacements()[0])

	// short ALLCAPS (<5 chars) still matched in Dutch
	matches = rule.Match(languagetool.AnalyzePlain("Mag ik die DVD lenen?"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "dvd", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Heb je de nieuwe IPAD al gezien?"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "iPad", matches[0].GetSuggestedReplacements()[0])
}
