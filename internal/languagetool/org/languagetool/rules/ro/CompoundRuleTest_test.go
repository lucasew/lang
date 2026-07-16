package ro

// Twin of languagetool-language-modules/ro/src/test/java/org/languagetool/rules/ro/CompoundRuleTest.java
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
	check(0, "Au plecat câteșitrei.")
	check(1, "câte și trei", "câteșitrei")
	check(1, "Câte și trei", "Câteșitrei")
	check(1, "câte-și-trei", "câteșitrei")
	check(1, "tus trei", "tustrei")
	check(1, "tus-trei", "tustrei")
}
