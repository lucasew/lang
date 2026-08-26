package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/UnitConversionRuleTest.java
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
	require.Equalf(t, expectedMatches, len(ms), "Got matches for %q: %v", input, summarizeDEUnitMatches(ms))
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

func summarizeDEUnitMatches(ms []*rules.RuleMatch) []string {
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
	// Java: Language lang = Languages.getLanguageForShortCode("de");
	//       JLanguageTool lt = new JLanguageTool(lang);
	//       UnitConversionRule rule = new UnitConversionRule(JLanguageTool.getMessageBundle(lang));
	rule := NewUnitConversionRule(nil)
	require.Equal(t, "EINHEITEN_METRISCH", rule.GetID())

	// Full Java UnitConversionRuleTest.match() matrix
	assertUnitConversionMatches(t, rule, "Ich bin 6 Fuß groß.", 1, "1,83 Meter")
	assertUnitConversionMatches(t, rule, "Ich bin 6 Fuß (2,02 m) groß.", 1, "1,83 Meter")
	assertUnitConversionMatches(t, rule, "Ich bin 6 Fuß (1,82 m) groß.", 0, "")
	// Java uses NBSP (\u00A0) between 1.800 and Pfund
	assertUnitConversionMatches(t, rule, "Der Kostenvoranschlag hatte eine Höhe von 1.800\u00A0Pfund Sterling.", 0, "")
	assertUnitConversionMatches(t, rule, "Der Weg ist 100 Meilen lang.", 1, "160,93 Kilometer")
	assertUnitConversionMatches(t, rule, "Der Weg ist 10 km (20 Meilen) lang.", 1, "6,21")
	assertUnitConversionMatches(t, rule, "Der Weg ist 10 km (6,21 Meilen) lang.", 0, "")
	assertUnitConversionMatches(t, rule, "Der Weg ist 100 Meilen (160,93 Kilometer) lang.", 0, "")
	assertUnitConversionMatches(t, rule, "Die Ladung ist 10.000,75 Pfund schwer.", 1, "4,54 Tonnen")
	assertUnitConversionMatches(t, rule, `Sie ist 5'6" groß.`, 1, "1,68 m")
	assertUnitConversionMatches(t, rule, "Meine neue Wohnung ist 500 sq ft groß.", 1, "46,45 Quadratmeter")
	// de-CH thousands must not be feet (Java antiPatterns + removeAntiPatternMatches)
	assertUnitConversionMatches(t, rule, "Zwischen 330'000 und 500'000/600", 0, "")
}

func TestUnitConversionRule_Meta(t *testing.T) {
	r := NewUnitConversionRule(nil)
	require.Equal(t, "EINHEITEN_METRISCH", r.GetID())
	require.Equal(t, "Schlägt vor oder überprüft Angaben des metrischen Äquivalentes bei bestimmten Maßeinheiten.", r.GetDescription())
	require.NotEmpty(t, r.GetIncorrectExamples())
}
