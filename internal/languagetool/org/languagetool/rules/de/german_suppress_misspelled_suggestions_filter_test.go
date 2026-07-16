package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGermanSuppressMisspelledSuggestionsFilter(t *testing.T) {
	f := NewGermanSuppressMisspelledSuggestionsFilter()
	kept, ok := f.FilterSuggestions([]string{"Haus", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"Haus", "xyz"}, kept)
	f.IsMisspelled = func(w string) bool { return w == "xyz" }
	kept, ok = f.FilterSuggestions([]string{"Haus", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"Haus"}, kept)
	_, ok = f.FilterSuggestions([]string{"xyz"}, true)
	require.False(t, ok)
}
