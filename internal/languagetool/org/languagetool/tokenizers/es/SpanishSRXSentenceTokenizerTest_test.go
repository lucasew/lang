package es

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpanishSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewSpanishSRXSentenceTokenizer()
	require.NotEmpty(t, tok.Tokenize("A. B."))
}
