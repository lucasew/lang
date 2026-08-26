package dev

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpandIrishNounFromGuess(t *testing.T) {
	// word ending in óir
	lines := ExpandIrishNounFromGuess("múinteoir") // ends with eoir? múinteoir ends eoir
	// "múinteoir" ends with "eoir"
	require.NotEmpty(t, lines)
	require.Equal(t, "múinteoir", lines[0].Lemma)
	// form = stem + ending form
	require.Contains(t, lines[0].Form, "eoir")
	require.Contains(t, lines[0].Tag, "Noun:")
}

func TestExpandIrishNounFromGuess_NoMatch(t *testing.T) {
	require.Empty(t, ExpandIrishNounFromGuess("cat"))
}
