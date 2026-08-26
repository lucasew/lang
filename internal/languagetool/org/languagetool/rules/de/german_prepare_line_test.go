package de

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrepareLineForSpeller(t *testing.T) {
	// Comment strip: Java split("#") keeps trailing space on form (no trim).
	require.Equal(t, []string{"Human-centered Design"}, PrepareLineForSpeller("Human-centered Design"))
	require.Equal(t, []string{"foo "}, PrepareLineForSpeller("foo # comment"))
	require.Equal(t, []string{"foo"}, PrepareLineForSpeller("foo#comment"))

	// /E /S /N expansions (Java formTag length == 2)
	require.Equal(t, []string{"Haus", "Hause"}, PrepareLineForSpeller("Haus/E"))
	require.Equal(t, []string{"Auto", "Autos"}, PrepareLineForSpeller("Auto/S"))
	require.Equal(t, []string{"Bar", "Barn"}, PrepareLineForSpeller("Bar/N"))
	require.Equal(t, []string{"Foo", "Fooe", "Foos", "Foon"}, PrepareLineForSpeller("Foo/ESN"))

	// Java: formTag.length != 2 → no expansions (only form)
	require.Equal(t, []string{"a"}, PrepareLineForSpeller("a/E/S"))

	// empty form + /E: Java still expands (form "" + "e")
	require.Equal(t, []string{"", "e"}, PrepareLineForSpeller("/E"))
}

func TestGermanMultitokenSpeller_PrepareLineWired(t *testing.T) {
	sp := NewGermanMultitokenSpeller()
	require.NotNil(t, sp.PrepareLine)
	// Load via PrepareLine expansion: multiword with /S should index both forms if multiword
	r := strings.NewReader("foo bar/S\n")
	require.NoError(t, sp.LoadWords(r))
	// getNormalizeKey multiword — "foo bar" and "foo bars" both registered
	// Suggestions for near-miss may be empty without full dict; just ensure load did not invent crash
	_ = sp.GetSuggestions("foo bar")
}
