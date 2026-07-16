package tokenizers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPersianSRXSentenceTokenizer_Test(t *testing.T) {
	// Persian wrapper lives in fa package; core SRX still works for smoke.
	tok := NewSRXSentenceTokenizer("fa")
	require.NotEmpty(t, tok.Tokenize("سلام. دنیا."))
}
