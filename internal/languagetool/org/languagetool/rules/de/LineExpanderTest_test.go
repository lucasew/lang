package de

// Twin of org.languagetool.rules.de.LineExpanderTest.testExpansion.
// Java assertSetsEqual → order-independent set equality.
// Verb-prefix cases use real GermanSynthesizer VER: forms (no invent form lists).

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLineExpander_Expansion(t *testing.T) {
	// Java: private final LineExpander exp = new LineExpander();
	// Verb cache is static GermanSynthesizer.INSTANCE — wire the same path.
	exp := WireLineExpander()
	require.NotNil(t, exp.VerbForms,
		"german_synth.dict required for LineExpander verb-prefix expansion (Java GermanSynthesizer.INSTANCE); install german-pos-dict / DE resources")

	// Java expand("") → singletonList("") (toString "[]" makes assertSetsEqual a no-op quirk).
	// Assert real list content: [""].
	require.Equal(t, []string{""}, exp.ExpandLine(""))

	assertSetsEqual(t, exp.ExpandLine("Das"), []string{"Das"})
	assertSetsEqual(t, exp.ExpandLine("Tisch/E"), []string{"Tisch", "Tische"})
	assertSetsEqual(t, exp.ExpandLine("Tische/N"), []string{"Tische", "Tischen"})
	assertSetsEqual(t, exp.ExpandLine("Auto/S"), []string{"Auto", "Autos"})
	assertSetsEqual(t, exp.ExpandLine("klein/A"), []string{"klein", "kleine", "kleiner", "kleines", "kleinen", "kleinem"})
	assertSetsEqual(t, exp.ExpandLine("x/NSE"), []string{"x", "xn", "xs", "xe"})
	assertSetsEqual(t, exp.ExpandLine("x/NA"), []string{"x", "xn", "xe", "xer", "xes", "xen", "xem"})
	assertSetsEqual(t, exp.ExpandLine("viertjüngste/A"), []string{"viertjüngste", "viertjüngster", "viertjüngstes", "viertjüngsten", "viertjüngstem"})
	assertSetsEqual(t, exp.ExpandLine("Das  #foo"), []string{"Das"})
	assertSetsEqual(t, exp.ExpandLine("Tisch/E  #bla #foo"), []string{"Tisch", "Tische"})
	assertSetsEqual(t, exp.ExpandLine("Goethestraße/T"), []string{"Goethestraße", "Goethestr."})
	assertSetsEqual(t, exp.ExpandLine("Goethestrasse/T"), []string{"Goethestrasse", "Goethestr."})
	assertSetsEqual(t, exp.ExpandLine("Zwingenberger Straße/T"), []string{"Zwingenberger Straße", "Zwingenberger Str."})
	assertSetsEqual(t, exp.ExpandLine("Zwingenberger Strasse/T"), []string{"Zwingenberger Strasse", "Zwingenberger Str."})

	assertSetsEqual(t, exp.ExpandLine(`Escape\/N`), []string{"Escape/N"})
	// Java: combination of escape and flag not supported yet (commented out)

	// Verb-prefix via real synthesizer (Java GermanSynthesizer.INSTANCE VER: forms)
	assertSetsEqual(t, exp.ExpandLine("rüber_machen  #bla #foo"), []string{
		"rübermach", "rübergemacht", "rübermachest", "rübermachst", "rübermache",
		"rübermachen", "rübermachet", "rübermachte", "rübermachend", "rübermachten",
		"rübermacht", "rübermachtest", "rübermachtet", "rüberzumachen", "Rübermachens",
	})
	assertSetsEqual(t, exp.ExpandLine("rüber_verschicken"), []string{
		"rüberverschickend", "rüberverschickst", "rüberverschick", "rüberverschickest",
		"rüberverschicktest", "rüberverschicke", "rüberverschicket", "rüberverschickte",
		"rüberverschicktet", "rüberverschickten", "rüberverschicken", "rüberverschickt",
		"rüberzuverschicken", "Rüberverschickens",
	})
	assertSetsEqual(t, exp.ExpandLine(`escape\_machen`), []string{"escape_machen"})

	// Java: expand("rüber/invalidword"); fail(); catch RuntimeException
	require.Panics(t, func() { exp.ExpandLine("rüber/invalidword") })
}

// assertSetsEqual ports LineExpanderTest.assertSetsEqual (order-independent).
func assertSetsEqual(t *testing.T, actual, expected []string) {
	t.Helper()
	require.ElementsMatch(t, expected, actual)
}
