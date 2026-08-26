package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatalanSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewCatalanSRXSentenceTokenizer()
	require.NotEmpty(t, tok.Tokenize("A. B."))
}
