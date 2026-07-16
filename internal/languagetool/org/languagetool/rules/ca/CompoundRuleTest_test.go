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
