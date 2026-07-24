package es

// Twin of languagetool-language-modules/es/src/test/java/org/languagetool/rules/es/CompoundRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCompoundRule_Rule(t *testing.T) {
	rule := NewCompoundRule(nil)
	check := func(expectedErrors int, text string, expSuggestions ...string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(text))
		require.Equal(t, expectedErrors, len(matches), "text %q got %v", text, matches)
		if len(expSuggestions) > 0 {
			require.Equal(t, 1, expectedErrors)
			require.Equal(t, expSuggestions, matches[0].GetSuggestedReplacements(), "text %q", text)
		}
	}

	// correct sentences:
	check(0, "Guinea-Conakri")

	// incorrect sentences:
	check(1, "Guinea Bisáu", "Guinea-Bisáu")
}

func TestCompoundRule_IsMisspelledViaTagIsTagged(t *testing.T) {
	rule := NewCompoundRule(nil)
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Guinea Bisáu"))))

	rule.TagIsTagged = func(word string) bool { return false }
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Guinea Bisáu"))))

	rule.TagIsTagged = func(word string) bool { return word == "Guinea-Bisáu" }
	matches := rule.Match(languagetool.AnalyzePlain("Guinea Bisáu"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Guinea-Bisáu"}, matches[0].GetSuggestedReplacements())
}
