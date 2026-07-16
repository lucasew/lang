package ngrams

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetContextStrings(t *testing.T) {
	tokens := []string{"a", "b", "c", "d"}
	ctx := GetContextStrings("b", tokens, "X", 1, 1)
	require.Equal(t, []string{"a", "X", "c"}, ctx)
	require.Panics(t, func() { GetContextStrings("z", tokens, "X", 1, 1) })
}

func TestGetContextGoogleTokens(t *testing.T) {
	tokens := []GoogleToken{
		NewGoogleToken("hello", 0, 5),
		NewGoogleToken("world", 6, 11),
		NewGoogleToken("test", 12, 16),
	}
	ctx := GetContextGoogleTokens(1, tokens, "earth", 1, 1)
	require.Equal(t, []string{"hello", "earth", "test"}, ctx)
}
