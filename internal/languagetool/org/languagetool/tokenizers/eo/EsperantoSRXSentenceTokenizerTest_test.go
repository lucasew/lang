package eo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEsperantoSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewEsperantoSRXSentenceTokenizer()
	require.NotEmpty(t, tok.Tokenize("A. B."))
}
