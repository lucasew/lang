package fr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMakeContractionsFilter(t *testing.T) {
	f := NewMakeContractionsFilter()
	require.Equal(t, "du pain", f.FixContractions("de le pain"))
	require.Equal(t, "au marché", f.FixContractions("à le marché"))
	require.Equal(t, "des enfants", f.FixContractions("de les enfants"))
	require.Equal(t, "aux enfants", f.FixContractions("à les enfants"))
	require.Equal(t, []string{"du chat", "au parc"}, f.MapSuggestions([]string{"de le chat", "à le parc"}))
}
