package tl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTagalogSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewTagalogSRXSentenceTokenizer()
	require.NotNil(t, tok)
	got := tok.Tokenize("Hello. World.")
	require.NotEmpty(t, got)
}
