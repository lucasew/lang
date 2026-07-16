package it

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestItalianSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewItalianSRXSentenceTokenizer()
	require.NotNil(t, tok)
	got := tok.Tokenize("Hello. World.")
	require.NotEmpty(t, got)
}
