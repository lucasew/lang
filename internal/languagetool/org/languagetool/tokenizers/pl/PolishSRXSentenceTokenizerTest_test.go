package pl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPolishSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewPolishSRXSentenceTokenizer()
	require.NotEmpty(t, tok.Tokenize("A. B."))
}
