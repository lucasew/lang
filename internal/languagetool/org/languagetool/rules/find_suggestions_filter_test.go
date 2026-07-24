package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindSuggestionsFilter_Collect(t *testing.T) {
	f := NewFindSuggestionsFilter()
	f.SpellingSuggestions = func(token string) []string {
		return []string{token, "casa", "caso", "cosa"}
	}
	f.MatchesDesiredPOS = func(c, pos string) bool {
		return c != "caso"
	}
	got := f.Collect("caza", "NCFS000", false, false)
	require.Equal(t, []string{"casa", "cosa"}, got)
}

func TestApplySuggestionTemplates(t *testing.T) {
	got := ApplySuggestionTemplates([]string{"el {suggestion}"}, []string{"cotxe"})
	require.Equal(t, []string{"el cotxe"}, got)
}
