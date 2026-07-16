package lt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLithuanianSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewLithuanianSRXSentenceTokenizer()
	require.NotNil(t, tok)
	got := tok.Tokenize("Labas. Pasaulis.")
	require.NotEmpty(t, got)
}
