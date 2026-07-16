package crh

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrimeanTatarSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewCrimeanTatarSRXSentenceTokenizer()
	require.NotNil(t, tok)
	got := tok.Tokenize("Hello. World.")
	require.NotEmpty(t, got)
}
