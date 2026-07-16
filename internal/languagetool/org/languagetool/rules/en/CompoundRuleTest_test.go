package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/CompoundRuleTest.java
import (
	"fmt"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestCompoundRule_Rule(t *testing.T) {
	rule := NewCompoundRule(nil)
	check := func(expectedErrors int, text string, expSuggestions ...string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(text))
		require.Equal(t, expectedErrors, len(matches), "text %q got %s", text, describeMatches(matches))
		if len(expSuggestions) > 0 {
			require.Equal(t, 1, expectedErrors, "suggestions only checked when expectedErrors==1")
			require.Equal(t, expSuggestions, matches[0].GetSuggestedReplacements(),
				"text %q suggestions", text)
		}
	}

	// correct sentences:
	check(0, "The software supports case-sensitive search.")
	check(0, "He is one-year-old.")
	check(0, "If they're educated people, they will know.")
	check(0, "Tiffany & Co chairman has to say something important")
	check(0, "Another one bites the dust")
	check(0, "We well received your email")
	check(0, "air-to-air")
	check(0, "non-party")
	check(0, "age-old")
	check(0, "able-bodied")
	check(0, "non-scientific")
	check(0, "This is the first ever green bond by a municipality.")
	check(0, "Semi Automatic") // desired?
	check(0, "Night Mare")     // desired?
	check(0, "This is a multi-module project.")

	// incorrect sentences:
	check(1, "case sensitive", "case-sensitive")
	check(1, "Young criminals must be re educated.")
	check(1, "And an other one bites the dust")
	check(1, "An other one bites the dust")

	check(1, "good-bye", "goodbye")
	check(1, "back-fire", "backfire")
	check(1, "back fire", "backfire")
	check(1, "air-to air", "air-to-air")
	check(1, "air to air", "air-to-air")
	check(1, "air to-air", "air-to-air")
	check(1, "air to -air", "air-to-air")
	check(1, "age old", "age-old")
	check(1, "able bodied", "able-bodied")
	check(1, "non scientific", "non-scientific", "nonscientific")
	check(1, "night-mare", "nightmare")
	check(1, "Night-mare", "Nightmare")
	check(1, "Night mare", "Nightmare")
	check(1, "semi automatic", "semi-automatic", "semiautomatic")
	check(1, "Semi automatic", "Semi-automatic", "Semiautomatic")
	check(1, "Dev-Ops", "DevOps")
	check(1, "Dev Ops", "DevOps")
	check(1, "Night-mare", "Nightmare")
	check(1, "Play Station", "PlayStation")
	check(1, "Play-Station", "PlayStation")
}

func describeMatches(matches []*rules.RuleMatch) string {
	var s string
	for i, m := range matches {
		if i > 0 {
			s += "; "
		}
		s += fmt.Sprintf("[%d:%d]%v", m.GetFromPos(), m.GetToPos(), m.GetSuggestedReplacements())
	}
	if s == "" {
		return "[]"
	}
	return s
}
