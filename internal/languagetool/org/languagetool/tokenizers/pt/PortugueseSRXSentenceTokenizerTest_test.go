package pt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPortugueseSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewPortugueseSRXSentenceTokenizer()
	require.NotNil(t, tok)
	got := tok.Tokenize("Hello. World.")
	require.NotEmpty(t, got)
}
