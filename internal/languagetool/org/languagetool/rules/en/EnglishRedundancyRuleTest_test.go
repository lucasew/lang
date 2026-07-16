package en

// Port of EnglishRedundancyRule example pairs (no dedicated Java unit test).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishRedundancyRule(t *testing.T) {
	rule := NewEnglishRedundancyRule(nil)

	// Example from Java rule: tuna fish → tuna
	matches := rule.Match(languagetool.AnalyzePlain("I ate tuna fish yesterday."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "tuna", matches[0].GetSuggestedReplacements()[0])

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("I ate tuna yesterday."))))

	matches = rule.Match(languagetool.AnalyzePlain("An added bonus for everyone."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "bonus", matches[0].GetSuggestedReplacements()[0])
}
