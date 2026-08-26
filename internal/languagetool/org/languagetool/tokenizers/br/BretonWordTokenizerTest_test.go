package br

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func tokStr(tokens []string) string {
	return "[" + strings.Join(tokens, ", ") + "]"
}

func TestBretonWordTokenizer_Tokenize(t *testing.T) {
	w := NewBretonWordTokenizer()
	tokens := w.Tokenize("Test c'h")
	require.Equal(t, 3, len(tokens))
	require.Equal(t, "[Test,  , c’h]", tokStr(tokens))

	tokens = w.Tokenize("Test c’h")
	require.Equal(t, 3, len(tokens))
	require.Equal(t, "[Test,  , c’h]", tokStr(tokens))

	tokens = w.Tokenize("C'hwerc'h merc'h gwerc'h war c'hwerc'h marc'h kalloc'h")
	require.Equal(t, 13, len(tokens))
	require.Equal(t, "[C’hwerc’h,  , merc’h,  , gwerc’h,  , war,  , c’hwerc’h,  , marc’h,  , kalloc’h]", tokStr(tokens))

	tokens2 := w.Tokenize("Test n’eo")
	require.Equal(t, 4, len(tokens2))
	require.Equal(t, "[Test,  , n’, eo]", tokStr(tokens2))
}
