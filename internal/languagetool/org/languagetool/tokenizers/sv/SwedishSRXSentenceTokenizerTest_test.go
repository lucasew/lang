package sv

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSwedishSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewSwedishSRXSentenceTokenizer()
	require.NotEmpty(t, tok.Tokenize("Hej. Världen."))
}
