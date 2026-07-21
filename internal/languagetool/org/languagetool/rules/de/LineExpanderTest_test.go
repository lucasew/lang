package de

// Twin of LineExpanderTest — suffix expansion; verb-prefix without synth emits join/zu/genitive.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLineExpander_Expansion(t *testing.T) {
	e := NewLineExpander()
	require.Equal(t, []string{""}, e.ExpandLine(""))
	require.Equal(t, []string{"Das"}, e.ExpandLine("Das"))
	require.Equal(t, []string{"Tisch", "Tische"}, e.ExpandLine("Tisch/E"))
	require.Equal(t, []string{"Tische", "Tischen"}, e.ExpandLine("Tische/N"))
	require.Equal(t, []string{"Auto", "Autos"}, e.ExpandLine("Auto/S"))
	got := e.ExpandLine("klein/A")
	require.ElementsMatch(t, []string{"klein", "kleine", "kleiner", "kleines", "kleinen", "kleinem"}, got)
	// multi-flag NSE
	got = e.ExpandLine("x/NSE")
	require.ElementsMatch(t, []string{"x", "xn", "xs", "xe"}, got)
	require.Equal(t, []string{"Das"}, e.ExpandLine("Das  #foo"))
	require.Equal(t, []string{"Tisch", "Tische"}, e.ExpandLine("Tisch/E  #bla #foo"))
	require.ElementsMatch(t, []string{"Goethestraße", "Goethestr."}, e.ExpandLine("Goethestraße/T"))
	require.ElementsMatch(t, []string{"Goethestrasse", "Goethestr."}, e.ExpandLine("Goethestrasse/T"))
	// escaped slash is not a flag
	require.Equal(t, []string{"Escape/N"}, e.ExpandLine(`Escape\/N`))
	// gender gap
	require.ElementsMatch(t, []string{
		"Lehrer_in", "Lehrer_innen", "Lehrer*in", "Lehrer*innen", "Lehrer:in", "Lehrer:innen",
	}, e.ExpandLine("Lehrer_in"))
	// verb prefix without synthesizer: fail-closed (no invent join/zu/genitive)
	require.Empty(t, e.ExpandLine("rüber_machen"))
	e.VerbForms = func(lemma string) []string {
		if lemma == "machen" {
			return []string{"machen", "machst"}
		}
		return nil
	}
	got = e.ExpandLine("rüber_machen")
	require.Contains(t, got, "rübermachen")
	require.Contains(t, got, "rübermachst")
	require.Contains(t, got, "rüberzumachen")
	require.Contains(t, got, "Rübermachens")
	// Character.isLowerCase(charAt(0)): skip non-lower forms
	e.VerbForms = func(string) []string { return []string{"Mach", "machen"} }
	got = e.ExpandLine("rüber_machen")
	require.NotContains(t, got, "rüberMach")
	require.Contains(t, got, "rübermachen")
	// escaped underscore is plain
	require.Equal(t, []string{"escape_machen"}, e.ExpandLine(`escape\_machen`))
	// unknown flag panics (Java RuntimeException)
	require.Panics(t, func() { e.ExpandLine("rüber/invalidword") })
	// empty synth forms panic (Java RuntimeException)
	e.VerbForms = func(string) []string { return nil }
	require.Panics(t, func() { e.ExpandLine("rüber_machen") })
}
