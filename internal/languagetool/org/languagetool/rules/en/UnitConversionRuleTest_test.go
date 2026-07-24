package en

// Twin of languagetool-language-modules/en/.../UnitConversionRuleTest.java
// Helper mirrors languagetool-core/.../UnitConversionRuleTestHelper.assertMatches:
// match count + first match's suggestions must contain the converted substring.
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// assertUnitConversionMatches ports UnitConversionRuleTestHelper.assertMatches
// (JLanguageTool.getAnalyzedSentence → rule.match; count + suggestion contains).
func assertUnitConversionMatches(t *testing.T, rule *UnitConversionRule, input string, expectedMatches int, converted string) {
	t.Helper()
	// Java: lt.getAnalyzedSentence(input) — plain analysis (no POS required by this rule).
	ms := rule.Match(languagetool.AnalyzePlain(input))
	require.Equalf(t, expectedMatches, len(ms), "Got matches for %q: %v", input, summarizeUnitMatches(ms))
	if expectedMatches > 0 && converted != "" {
		require.NotEmpty(t, ms)
		suggestionCorrect := false
		var suggestion string
		for _, s := range ms[0].GetSuggestedReplacements() {
			if strings.Contains(s, converted) {
				suggestionCorrect = true
				suggestion = s
				break
			}
		}
		require.Truef(t, suggestionCorrect, "Suggestion is correct: %q / expected contains: %q (all: %v)",
			suggestion, converted, ms[0].GetSuggestedReplacements())
	}
}

func summarizeUnitMatches(ms []*rules.RuleMatch) []string {
	out := make([]string, 0, len(ms))
	for _, m := range ms {
		if m == nil {
			out = append(out, "<nil>")
			continue
		}
		out = append(out, m.GetMessage()+" | "+strings.Join(m.GetSuggestedReplacements(), "; "))
	}
	return out
}

func TestUnitConversionRule_Match(t *testing.T) {
	// Java: Language lang = Languages.getLanguageForShortCode("en");
	//       JLanguageTool lt = new JLanguageTool(lang);
	//       UnitConversionRule rule = new UnitConversionRule(JLanguageTool.getMessageBundle(lang));
	rule := NewUnitConversionRule(nil)

	// Full Java UnitConversionRuleTest.match() matrix
	assertUnitConversionMatches(t, rule, "I am 6 feet tall.", 1, "1.83 m")
	assertUnitConversionMatches(t, rule, "I am 6 feet (2.02 m) tall.", 1, "1.83 m")
	assertUnitConversionMatches(t, rule, "I am 6 feet (1.82 m) tall.", 0, "")
	assertUnitConversionMatches(t, rule, "Heat up to 18\u00A0°C (64.4 °F).", 0, "")
	assertUnitConversionMatches(t, rule, "Heat up to 18\u00A0°C (64.4\u00A0°F).", 0, "")
	assertUnitConversionMatches(t, rule, "Heat up to 18 °C (64.4\u00A0°F).", 0, "")
	assertUnitConversionMatches(t, rule, "The path is 100 miles long.", 1, "160.93 km")
	assertUnitConversionMatches(t, rule, "The path is 100 miles (160.93 km) long.", 0, "")
	assertUnitConversionMatches(t, rule, "The shipment weighs 10,000.75 pounds.", 1, "4.54 t")
	assertUnitConversionMatches(t, rule, "She is 5'6\".", 1, "1.68 m")
	assertUnitConversionMatches(t, rule, "My new apartment is 500 sq ft.", 1, "46.45 m²")
	assertUnitConversionMatches(t, rule, "It is 100 degrees Fahrenheit outside.", 1, "37.78 °C")
	assertUnitConversionMatches(t, rule, "It is 100 °F outside.", 1, "37.78 °C")
	assertUnitConversionMatches(t, rule, "It is -22 °F outside.", 1, "-30 °C")
	assertUnitConversionMatches(t, rule, "It is -22 degrees Fahrenheit outside.", 1, "-30 °C")

	// https://github.com/languagetool-org/languagetool/issues/2357
	assertUnitConversionMatches(t, rule, "Millions watched the 1989's Superbowl.", 0, "")

	// https://github.com/languagetool-org/languagetool/issues/2255
	assertUnitConversionMatches(t, rule, "Coordinates: 34°09′22″N 118°7′55″W", 0, "")
}
