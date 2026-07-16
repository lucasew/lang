package de

// Twin of GermanNumberInWordFilterTest — surface digit-in-word suggestions.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGermanNumberInWordFilter(t *testing.T) {
	f := NewGermanNumberInWordFilter()
	sugg := f.Suggestions("H0use")
	require.Contains(t, sugg, "House")
	sugg2 := f.Suggestions("H4us")
	require.Contains(t, sugg2, "Hus")
	require.Empty(t, f.Suggestions("Haus"))
}
