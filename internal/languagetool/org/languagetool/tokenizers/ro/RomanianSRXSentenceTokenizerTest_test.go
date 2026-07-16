package ro

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRomanianSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewRomanianSRXSentenceTokenizer()
	require.NotEmpty(t, tok.Tokenize("A. B."))
}
