package el

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGreekSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewGreekSRXSentenceTokenizer()
	require.NotEmpty(t, tok.Tokenize("A. B."))
}
