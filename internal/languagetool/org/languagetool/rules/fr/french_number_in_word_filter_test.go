package fr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFrenchNumberInWordFilter(t *testing.T) {
	f := NewFrenchNumberInWordFilter()
	require.Equal(t, []string{"mot", "mt"}, f.Suggestions("m0t"))
}

func TestFrenchSuppressMisspelled(t *testing.T) {
	f := NewFrenchSuppressMisspelledSuggestionsFilter()
	kept, ok := f.FilterSuggestions([]string{"bon"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"bon"}, kept)
}
