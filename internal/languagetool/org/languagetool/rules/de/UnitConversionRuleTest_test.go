package de

// Twin of UnitConversionRuleTest (simplified surface conversions).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUnitConversionRule_Match(t *testing.T) {
	rule := NewUnitConversionRule(nil)
	matchN := func(s string) int {
		return len(rule.Match(languagetool.AnalyzePlain(s)))
	}
	require.Equal(t, 1, matchN("Ich bin 6 Fuß groß."))
	require.Equal(t, 1, matchN("Der Weg ist 100 Meilen lang."))
	// already has metric nearby
	require.Equal(t, 0, matchN("Ich bin 6 Fuß (1,82 m) groß."))
	require.Equal(t, 0, matchN("Der Kostenvoranschlag hatte eine Höhe von 1.800 Pfund Sterling."))
}
