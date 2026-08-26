package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdaptSuggestionsFilter(t *testing.T) {
	f := NewAdaptSuggestionsFilter(func(s, orig string) string {
		return strings.ToUpper(s)
	})
	require.Equal(t, []string{"FOO", "BAR"}, f.MapSuggestions([]string{"foo", "bar"}, ""))
	// identity default
	f2 := NewAdaptSuggestionsFilter(nil)
	require.Equal(t, []string{"x"}, f2.MapSuggestions([]string{"x"}, "y"))
}
