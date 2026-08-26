package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuppressMisspelledSuggestionsFilter(t *testing.T) {
	f := NewSuppressMisspelledSuggestionsFilter()
	kept, ok := f.FilterSuggestions([]string{"a", "b"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"a", "b"}, kept)

	f.IsMisspelled = func(w string) bool { return w == "b" }
	kept, ok = f.FilterSuggestions([]string{"a", "b"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"a"}, kept)

	kept, ok = f.FilterSuggestions([]string{"b"}, true)
	require.False(t, ok)
	require.Nil(t, kept)

	kept, ok = f.FilterSuggestions([]string{"b"}, false)
	require.True(t, ok)
	require.Nil(t, kept)
}
