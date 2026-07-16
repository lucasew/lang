package multitoken

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultitokenSpeller_Exact(t *testing.T) {
	m := NewMultitokenSpeller()
	require.NoError(t, m.LoadWords(strings.NewReader(`# c
New York
Los Angeles
`)))
	// exact match → stop (no suggestions)
	require.Empty(t, m.GetSuggestions("New York"))
	// diacritic/case normalize exact via no-spaces? "new york" lower equals dict after normalize
	// original "new york" != "New York" but stopSearching title case of lower candidate:
	// candidate "New York" is not all-lower; second loop checks candidate==toLower → false
	// so we get suggestions from noSpaces key
	sugg := m.GetSuggestions("new york")
	require.Contains(t, sugg, "New York")
}

func TestMultitokenSpeller_Typo(t *testing.T) {
	m := NewMultitokenSpeller()
	require.NoError(t, m.LoadWords(strings.NewReader("hello world\n")))
	sugg := m.GetSuggestions("helo world")
	require.Contains(t, sugg, "hello world")
}

func TestGetNormalizeKey(t *testing.T) {
	require.Equal(t, "cafe latte", getNormalizeKey("Café-Latte"))
}
