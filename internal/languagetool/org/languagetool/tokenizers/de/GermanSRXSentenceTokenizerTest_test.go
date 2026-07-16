package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGermanSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewGermanSRXSentenceTokenizer()
	require.NotNil(t, tok)
	got := tok.Tokenize("Hello. World.")
	require.NotEmpty(t, got)
}
