package pl

// Twin of languagetool-language-modules/pl/src/test/java/org/languagetool/rules/pl/DashRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDashRule_Rule(t *testing.T) {
	rule := NewDashRule(nil)
	check := func(expectedErrors int, text string, expSuggestions ...string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(text))
		require.Equal(t, expectedErrors, len(matches), "text %q got %v", text, matches)
		if len(expSuggestions) > 0 {
			require.Equal(t, 1, expectedErrors)
			require.Equal(t, expSuggestions, matches[0].GetSuggestedReplacements(), "text %q", text)
		}
	}
	check(0, "Nie róbmy nic na łapu-capu.")
	check(0, "Jedzmy kogel-mogel.")
	check(0, "To jest ładna nota — bene, bene — odpowiedział Józek.")
	check(1, "bim – bom", "bim-bom")
	check(1, "Papua–Nowa Gwinea", "Papua-Nowa")
	check(1, "Papua — Nowa Gwinea", "Papua-Nowa")
	check(1, "Aix — en — Provence", "Aix-en-Provence")
}
