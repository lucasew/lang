package sr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSerbianSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewSerbianSRXSentenceTokenizer()
	require.NotEmpty(t, tok.Tokenize("Zdravo. Svete."))
}
