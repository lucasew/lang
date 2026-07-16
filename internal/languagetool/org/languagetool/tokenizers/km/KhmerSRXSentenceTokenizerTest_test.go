package km

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKhmerSRXSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewKhmerSRXSentenceTokenizer()
	require.NotEmpty(t, tok.Tokenize("A. B."))
}
