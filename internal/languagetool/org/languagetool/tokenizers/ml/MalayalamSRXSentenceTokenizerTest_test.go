package ml

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMalayalamSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewMalayalamSRXSentenceTokenizer()
	require.NotNil(t, tok)
	got := tok.Tokenize("Hello. World.")
	require.NotEmpty(t, got)
}
