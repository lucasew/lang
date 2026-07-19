package de

// Twin of UnitConversionRuleTest (simplified surface conversions).
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUnitConversionRule_Match(t *testing.T) {
	rule := NewUnitConversionRule(nil)
	require.Equal(t, "EINHEITEN_METRISCH", rule.GetID())
	matchN := func(s string) int {
		return len(rule.Match(languagetool.AnalyzePlain(s)))
	}
	require.Equal(t, 1, matchN("Ich bin 6 Fuß groß."))
	require.Equal(t, 1, matchN("Der Weg ist 100 Meilen lang."))
	// already has metric nearby
	require.Equal(t, 0, matchN("Ich bin 6 Fuß (1,82 m) groß."))
	require.Equal(t, 0, matchN("Der Kostenvoranschlag hatte eine Höhe von 1.800 Pfund Sterling."))
	// DE thousands + Pfund mass
	ms := rule.Match(languagetool.AnalyzePlain("Die Ladung ist 10.000,75 Pfund schwer."))
	require.NotEmpty(t, ms)
	require.Contains(t, strings.Join(ms[0].GetSuggestedReplacements(), " "), "Tonnen")

	// specialPatterns: 5'6" / 5ft 6 → feet+inches (Java AbstractUnitConversionRule)
	require.Equal(t, 1, matchN(`Er ist 5'6" groß.`))
	require.Equal(t, 1, matchN("Er ist 5ft 6in groß."))
	// already metric nearby
	require.Equal(t, 0, matchN(`Er ist 5'6" (1,68 m) groß.`))

	// CHECK path: wrong parenthetical conversion (Java Message.CHECK)
	msCheck := rule.Match(languagetool.AnalyzePlain("Ich bin 6 Fuß (3 m) groß."))
	require.NotEmpty(t, msCheck)
	// should flag the conversion paren, not only suggest missing metric
	require.Contains(t, msCheck[0].GetMessage(), "falsch")
}
