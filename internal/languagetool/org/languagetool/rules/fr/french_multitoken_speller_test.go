package fr

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFrenchMultitokenSpeller(t *testing.T) {
	require.Equal(t, []string{
		"/fr/multiwords.txt",
		"/spelling_global.txt",
		"/fr/hyphenated_words.txt",
	}, FrenchMultitokenResourcePaths)
	s := NewFrenchMultitokenSpeller()
	require.NoError(t, s.LoadWords(strings.NewReader("New York\n")))
	require.Contains(t, s.GetSuggestions("new york"), "New York")
	require.NotNil(t, FrenchMultitokenSpellerInstance)
}
