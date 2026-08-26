package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/rules/ca/CompoundRuleTest.java
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
			require.Equal(t, expSuggestions, matches[0].GetSuggestedReplacements())
		}
	}
	check(0, "tam-tam")
	check(1, "Ryan-Air", "Ryanair")
}

func TestCompoundRule_IsMisspelledViaTagIsTagged(t *testing.T) {
	rule := NewCompoundRule(nil)
	// default: keep suggestion
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Ryan-Air"))))

	rule.TagIsTagged = func(word string) bool { return false }
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ryan-Air"))), "untagged form drops suggestion")

	rule.TagIsTagged = func(word string) bool { return word == "Ryanair" }
	matches := rule.Match(languagetool.AnalyzePlain("Ryan-Air"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Ryanair"}, matches[0].GetSuggestedReplacements())
}
