package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/SwissCompoundRuleTest.java
// Also runs shared GermanCompoundRuleTest cases via runDECompoundTests (like Java AbstractCompoundRuleTest).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSwissCompoundRule_Rule(t *testing.T) {
	rule := NewSwissCompoundRule(nil)
	require.Equal(t, "DE_CH_COMPOUNDS", rule.GetID())

	// Java SwissCompoundRuleTest (ß forms; süss commented out in Java but expander covers ss)
	check := func(expected int, text string) {
		t.Helper()
		ms := rule.Match(languagetool.AnalyzePlain(text))
		require.Equal(t, expected, len(ms), "text %q", text)
	}
	check(1, "Ente süß-sauer")
	check(1, "Ente süß sauer")
	// SwissExpandLine: ß line also matches ss surface
	check(1, "Ente süss-sauer")
	check(1, "Ente süss sauer")

	// expander twin
	require.Equal(t, []string{"süß-sauer*", "süss-sauer*"}, SwissExpandLine("süß-sauer*"))
	require.Equal(t, []string{"no-beta"}, SwissExpandLine("no-beta"))
}
