package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Unit checks for LineExpander flag paths and fail-closed verb-prefix without synth.
// Full twin of LineExpanderTest lives in LineExpanderTest_test.go (real synthesizer).

func TestLineExpander(t *testing.T) {
	e := NewLineExpander()
	require.Equal(t, []string{"foo", "foos"}, e.ExpandLine("foo/S"))
	require.Equal(t, []string{"bar", "barn"}, e.ExpandLine("bar/N"))
	// Without VerbForms (synth not wired): fail-closed — no invent join/zu/genitive
	require.Empty(t, e.ExpandLine("weiter_gehen"))
	require.Equal(t, []string{"plain"}, e.ExpandLine("plain"))

	// Multi-slash like Java split → IllegalArgumentException
	require.Panics(t, func() { e.ExpandLine("a/b/c") })
	// Empty synth forms panic (Java RuntimeException)
	e.VerbForms = func(string) []string { return nil }
	require.Panics(t, func() { e.ExpandLine("rüber_machen") })
}

func TestLineExpander_GenderGap(t *testing.T) {
	e := NewLineExpander()
	// Java handleLineWithPrefix parts[1].equals("in") special case (not in LineExpanderTest)
	require.Equal(t, []string{
		"Lehrer_in", "Lehrer_innen", "Lehrer*in", "Lehrer*innen", "Lehrer:in", "Lehrer:innen",
	}, e.ExpandLine("Lehrer_in"))
}

func TestLineExpander_FFlag(t *testing.T) {
	e := NewLineExpander()
	// /F → word + "in" (m/f); not covered by LineExpanderTest but in Java handleLineWithFlags
	require.Equal(t, []string{"Arzt", "Arztin"}, e.ExpandLine("Arzt/F"))
}
