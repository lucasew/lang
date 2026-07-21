package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLineExpander(t *testing.T) {
	e := NewLineExpander()
	require.Equal(t, []string{"foo", "foos"}, e.ExpandLine("foo/S"))
	require.Equal(t, []string{"bar", "barn"}, e.ExpandLine("bar/N"))
	// Without VerbForms: Java synthesizer always present; empty forms → RuntimeException
	require.Panics(t, func() { e.ExpandLine("weiter_gehen") })
	// With synth-like VerbForms including infinitive:
	e.VerbForms = func(lemma string) []string {
		if lemma == "gehen" {
			return []string{"gehen", "geht"}
		}
		return nil
	}
	got := e.ExpandLine("weiter_gehen")
	require.Contains(t, got, "weitergehen")
	require.Contains(t, got, "weitergeht")
	require.Contains(t, got, "weiterzugehen")
	require.Contains(t, got, "Weitergehens")
	require.Equal(t, []string{"plain"}, e.ExpandLine("plain"))
}
