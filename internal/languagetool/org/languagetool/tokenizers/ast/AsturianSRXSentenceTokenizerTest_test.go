package ast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAsturianSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewAsturianSRXSentenceTokenizer()
	require.NotNil(t, tok)
	got := tok.Tokenize("Hola. Mundu.")
	require.NotEmpty(t, got)
}
