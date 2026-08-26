package es

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpanishMultitokenSpeller(t *testing.T) {
	require.Equal(t, []string{
		"/es/multiwords.txt",
		"/spelling_global.txt",
		"/es/hyphenated_words.txt",
	}, SpanishMultitokenResourcePaths)
	s := NewSpanishMultitokenSpeller()
	require.NoError(t, s.LoadWords(strings.NewReader("Nueva York\n")))
	require.Contains(t, s.GetSuggestions("nueva york"), "Nueva York")
	require.NotNil(t, SpanishMultitokenSpellerInstance)
}
