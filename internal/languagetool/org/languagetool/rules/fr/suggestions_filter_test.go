package fr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuggestionsFilter(t *testing.T) {
	f := NewSuggestionsFilter()
	got := f.Filter([]string{"bonjour", "xyz123", "salut"}, `.*\d.*`)
	require.Equal(t, []string{"bonjour", "salut"}, got)
}
