package nl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDutchSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewDutchSRXSentenceTokenizer()
	require.NotNil(t, tok)
	got := tok.Tokenize("Dit is een zin. En nog een.")
	require.NotEmpty(t, got)
}
