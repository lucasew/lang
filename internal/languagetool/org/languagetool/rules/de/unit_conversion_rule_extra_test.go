package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestUnitConversionRule_FullJavaUnitsRegistered(t *testing.T) {
	r := NewUnitConversionRule(nil)
	require.Equal(t, "EINHEITEN_METRISCH", r.GetID())
	// Imperial still suggests metric
	ms := r.Match(languagetool.AnalyzePlain("Der Weg ist 10 Meilen lang."))
	require.NotEmpty(t, ms)
	// Cubic metric names registered (Match must not panic on Kubik surface)
	_ = r.Match(languagetool.AnalyzePlain("Das Volumen ist 2 Kubikmeter."))
	// Java getMessage SUGGESTION
	require.Contains(t, r.GetMessage(rules.UnitMsgSuggestion), "metrische")
	// Java getShortMessage SUGGESTION
	require.Equal(t, "Metrisches Äquivalent hinzufügen?", r.GetShortMessage(rules.UnitMsgSuggestion))
	require.NotEmpty(t, ms[0].GetShortMessage())
}

// Java UnitConversionRule ctor: addExamplePair(6 Fuß → 6 Fuß (1,83 m)).
func TestUnitConversionRule_ExamplePair(t *testing.T) {
	r := NewUnitConversionRule(nil)
	inc := r.GetIncorrectExamples()
	cor := r.GetCorrectExamples()
	require.Len(t, inc, 1)
	require.Len(t, cor, 1)
	require.Equal(t, "Ich bin <marker>6 Fuß</marker> groß.", inc[0].GetExample())
	require.Equal(t, []string{"6 Fuß (1,83 m)"}, inc[0].GetCorrections())
	require.Equal(t, "Ich bin <marker>6 Fuß (1,83 m)</marker> groß.", cor[0].GetExample())
}
