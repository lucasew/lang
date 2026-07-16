package fr

// Twin of languagetool-language-modules/fr/src/test/java/org/languagetool/rules/fr/CompoundRuleTest.java
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
	check(0, "Jésus-Christ")
	check(0, "Congo-Brazzaville")
	check(0, "vidéo-clip")
	check(0, "anglo-saxon")

	// incorrect sentences:
	check(1, "Jésus Christ")
	check(1, "Congo Brazzaville")
	check(1, "Congo- Brazzaville")
	check(1, "Congo -Brazzaville")

	check(1, "rez-de chaussée", "rez-de-chaussée")
	check(1, "Congo -Brazzaville", "Congo-Brazzaville")
	check(1, "Congo- Brazzaville", "Congo-Brazzaville")
	check(1, "Congo - Brazzaville", "Congo-Brazzaville")

	check(1, "le - quel", "lequel")
	check(1, "le quel", "lequel")
	check(1, "le- quel", "lequel")

	check(1, "anglo saxon", "anglo-saxon")
	check(1, "anglo- saxon", "anglo-saxon")
	check(1, "anglo -saxon", "anglo-saxon")
	check(1, "anglo - saxon", "anglo-saxon")
}
