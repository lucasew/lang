package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindSuggestionsEsFilter(t *testing.T) {
	f := NewFindSuggestionsEsFilter()
	got := f.RewriteEsSuggestions([]struct{ Form, POS string }{
		{"casa", "NCFS000"},
		{"canta", "VMIP3S0"},
		{"foo", "XXXX"},
	}, 10)
	require.Equal(t, []string{"és casa", "es canta"}, got)
}
