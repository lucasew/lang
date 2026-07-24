package gl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGalicianSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewGalicianSRXSentenceTokenizer()
	require.NotEmpty(t, tok.Tokenize("A. B."))
}
