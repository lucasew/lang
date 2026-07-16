package sv

// Twin of languagetool-language-modules/sv/src/test/java/org/languagetool/rules/sv/CompoundRuleTest.java
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
	check(0, "IP-Adress")
	check(0, "moll-tonart")
	check(0, "e-mail")
	check(1, "skit bra", "skitbra")
	check(1, "skit-bra", "skitbra")
	check(1, "IP Adress", "IP-Adress")
	check(1, "moll tonart", "moll-tonart", "molltonart")
	check(1, "e mail", "e-mail")
}
