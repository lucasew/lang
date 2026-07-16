package de

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Unit tests for GermanWordTokenizer (no dedicated Java WordTokenizerTest).

func tokStr(t []string) string { return "[" + strings.Join(t, ", ") + "]" }

func TestGermanWordTokenizer_Tokenize(t *testing.T) {
	w := NewGermanWordTokenizer()
	// underscore is a delimiter
	tokens := w.Tokenize("foo_bar")
	require.Equal(t, "[foo, _, bar]", tokStr(tokens))
	// low-9 quotation mark ‚ is a delimiter
	tokens = w.Tokenize("sagte‚hallo")
	require.Equal(t, "[sagte, ‚, hallo]", tokStr(tokens))
	// basic whitespace
	tokens = w.Tokenize("Das ist\u00A0ein Test")
	require.Equal(t, "[Das,  , ist, \u00A0, ein,  , Test]", tokStr(tokens))
}
