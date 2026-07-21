package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/UnitConversionRuleTest.java
// via UnitConversionRuleTestHelper.assertMatches (contains converted substring).
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUnitConversionRule_Match(t *testing.T) {
	rule := NewUnitConversionRule(nil)
	require.Equal(t, "EINHEITEN_METRISCH", rule.GetID())

	assertMatches := func(input string, expectedN int, converted *string) {
		t.Helper()
		ms := rule.Match(languagetool.AnalyzePlain(input))
		require.Equal(t, expectedN, len(ms), "input=%q", input)
		if expectedN > 0 && converted != nil {
			ok := false
			for _, s := range ms[0].GetSuggestedReplacements() {
				if strings.Contains(s, *converted) {
					ok = true
					break
				}
			}
			require.True(t, ok, "want suggestion containing %q, got %v for %q",
				*converted, ms[0].GetSuggestedReplacements(), input)
		}
	}
	strp := func(s string) *string { return &s }

	// Java UnitConversionRuleTest#match
	assertMatches("Ich bin 6 Fuß groß.", 1, strp("1,83 Meter"))
	assertMatches("Ich bin 6 Fuß (2,02 m) groß.", 1, strp("1,83 Meter"))
	assertMatches("Ich bin 6 Fuß (1,82 m) groß.", 0, nil)
	assertMatches("Der Kostenvoranschlag hatte eine Höhe von 1.800 Pfund Sterling.", 0, nil)
	// NBSP may be regular space in AnalyzePlain — also try plain space
	assertMatches("Der Kostenvoranschlag hatte eine Höhe von 1.800 Pfund Sterling.", 0, nil)
	assertMatches("Der Weg ist 100 Meilen lang.", 1, strp("160,93 Kilometer"))
	assertMatches("Der Weg ist 10 km (20 Meilen) lang.", 1, strp("6,21"))
	assertMatches("Der Weg ist 10 km (6,21 Meilen) lang.", 0, nil)
	assertMatches("Der Weg ist 100 Meilen (160,93 Kilometer) lang.", 0, nil)
	assertMatches("Die Ladung ist 10.000,75 Pfund schwer.", 1, strp("4,54 Tonnen"))
	assertMatches(`Sie ist 5'6" groß.`, 1, strp("1,68 m"))
	assertMatches("Meine neue Wohnung ist 500 sq ft groß.", 1, strp("46,45 Quadratmeter"))
	// de-CH thousands must not be feet (Java antiPatterns + removeAntiPatternMatches)
	assertMatches("Zwischen 330'000 und 500'000/600", 0, nil)

	// CHECK path message for wrong parenthetical
	msCheck := rule.Match(languagetool.AnalyzePlain("Ich bin 6 Fuß (3 m) groß."))
	require.NotEmpty(t, msCheck)
	require.Contains(t, msCheck[0].GetMessage(), "falsch")
}

func TestUnitConversionRule_Meta(t *testing.T) {
	r := NewUnitConversionRule(nil)
	require.Equal(t, "EINHEITEN_METRISCH", r.GetID())
	require.Contains(t, r.GetDescription(), "metrischen")
	require.NotEmpty(t, r.GetIncorrectExamples())
}
