package pt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPortugueseMultitokenSpeller(t *testing.T) {
	require.Equal(t, []string{
		"/pt/multiwords.txt",
		"/spelling_global.txt",
		"/pt/hyphenated_words.txt",
	}, PortugueseMultitokenResourcePaths)
	s := NewPortugueseMultitokenSpeller()
	require.NoError(t, s.LoadWords(strings.NewReader("São Paulo\n")))
	require.Contains(t, s.GetSuggestions("sao paulo"), "São Paulo")
	require.NotNil(t, PortugueseMultitokenSpellerInstance)
}
