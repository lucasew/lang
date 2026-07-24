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

func TestCompoundRule_IsMisspelledViaTagIsTagged(t *testing.T) {
	rule := NewCompoundRule(nil)
	// Without TagIsTagged, hyphen suggestions kept (default isMisspelled false).
	matches := rule.Match(languagetool.AnalyzePlain("anglo saxon"))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetSuggestedReplacements(), "anglo-saxon")

	// When TagIsTagged says the hyphenated form is NOT tagged → drop suggestion.
	rule.TagIsTagged = func(word string) bool {
		return false // nothing tagged → isMisspelled true → filter drops all
	}
	matches = rule.Match(languagetool.AnalyzePlain("anglo saxon"))
	// filterReplacements empties → no match emitted (or empty replacements path)
	// AbstractCompoundRule: if len(replacement)==0 break without adding match
	require.Equal(t, 0, len(matches), "untagged suggestions fail closed")

	// When TagIsTagged says form is tagged → keep
	rule.TagIsTagged = func(word string) bool {
		return word == "anglo-saxon"
	}
	matches = rule.Match(languagetool.AnalyzePlain("anglo saxon"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"anglo-saxon"}, matches[0].GetSuggestedReplacements())
}

func TestWireCompoundRuleTagger(t *testing.T) {
	rule := NewCompoundRule(nil)
	WireCompoundRuleTagger(rule, func(word string) []*languagetool.AnalyzedTokenReadings {
		if word == "anglo-saxon" {
			pos := "N"
			return []*languagetool.AnalyzedTokenReadings{
				languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(word, &pos, nil)),
			}
		}
		return []*languagetool.AnalyzedTokenReadings{
			languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(word, nil, nil)),
		}
	})
	require.True(t, rule.TagIsTagged("anglo-saxon"))
	require.False(t, rule.TagIsTagged("xyzzy-not-a-word"))
}
