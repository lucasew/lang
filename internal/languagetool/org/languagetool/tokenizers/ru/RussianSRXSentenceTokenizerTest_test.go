package ru

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRussianSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewRussianSRXSentenceTokenizer()
	require.NotNil(t, tok)
	got := tok.Tokenize("Hello. World.")
	require.NotEmpty(t, got)
}
