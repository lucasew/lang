package nl

// Twin of languagetool-language-modules/nl/src/test/java/org/languagetool/rules/nl/SimpleReplaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("all right"))))
	// no match b/c case-sensitivity (dictionary key is lowercase "kudde eigenschappen")
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("De Kudde eigenschappen"))))

	check := func(sentence, suggestion string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(sentence))
		require.Equal(t, 1, len(matches), "sentence %q", sentence)
		require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
		require.Equal(t, suggestion, matches[0].GetSuggestedReplacements()[0], "sentence %q", sentence)
	}
	// sentence-start capitalisation of suggestion
	check("klaa", "Klaar")
	check("een BTW nummer", "btw-nummer")
	check("kleurweergave eigenschappen.", "Kleurweergave-eigenschappen")
	check("De kleurweergave eigenschappen.", "kleurweergave-eigenschappen")
}
