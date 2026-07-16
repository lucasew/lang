package fa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPersianSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewPersianSRXSentenceTokenizer()
	require.NotEmpty(t, tok.Tokenize("A. B."))
}
