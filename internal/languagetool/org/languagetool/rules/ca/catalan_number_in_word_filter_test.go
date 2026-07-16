package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatalanNumberInWordFilter(t *testing.T) {
	f := NewCatalanNumberInWordFilter()
	require.Equal(t, []string{"cas"}, f.Suggestions("cas4"))
}

func TestCatalanSuppressMisspelled(t *testing.T) {
	f := NewCatalanSuppressMisspelledSuggestionsFilter()
	f.IsMisspelled = func(w string) bool { return w == "xyz" }
	kept, ok := f.FilterSuggestions([]string{"bé", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"bé"}, kept)
}
