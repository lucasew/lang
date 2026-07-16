package ru

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port-style unit tests for RussianWordTokenizer (no dedicated Java *WordTokenizerTest).

func tokStr(tokens []string) string {
	return "[" + strings.Join(tokens, ", ") + "]"
}

func TestRussianWordTokenizer_Tokenize(t *testing.T) {
	w := NewRussianWordTokenizer()
	// abbreviations with slash stay whole
	tokens := w.Tokenize("купить б/у телефон")
	require.Contains(t, tokens, "б/у")
	tokens = w.Tokenize("оплата б/н")
	require.Contains(t, tokens, "б/н")
	// apostrophe and period are delimiters
	tokens = w.Tokenize("кто-то")
	// hyphen may still be base delimiter
	require.NotEmpty(t, tokens)
	tokens = w.Tokenize("слово.слово")
	require.Equal(t, "[слово, ., слово]", tokStr(tokens))
	// protected " ."
	tokens = w.Tokenize("конец .")
	require.Contains(t, tokens, ".")
}
