package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishDashRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishDashRule_Rule(t *testing.T) {
	rule := NewEnglishDashRule(nil)
	check := func(expectedErrors int, text string, expectedSuggestion string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(text))
		require.Equal(t, expectedErrors, len(matches), "text %q got %v", text, matches)
		if expectedSuggestion != "" && len(matches) > 0 {
			require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
			require.Equal(t, expectedSuggestion, matches[0].GetSuggestedReplacements()[0])
		}
	}

	check(0, "This is my T-shirt.", "")
	check(0, "This is water-proof.", "")
	check(0, "This works semi-automatically.", "")
	check(0, "She's a newcomer.", "")
	check(0, "I sent you and e-mail.", "")

	check(1, "T – shirt", "T-shirt")
	check(1, "three–way street", "three-way")
	check(1, "surface — to — surface", "surface-to-surface")
	check(1, "This works semi–automatically.", "semi-automatically")
	check(1, "This works semi – automatically.", "semi-automatically")
	check(1, "I sent you and e–mail.", "e-mail")
}
