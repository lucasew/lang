package da

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDanishSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewDanishSRXSentenceTokenizer()
	require.NotEmpty(t, tok.Tokenize("Hej. Verden."))
}
