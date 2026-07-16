package pl

// Twin of PolishSentenceTokenizerTest — SRX green smokes.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPolishSentenceTokenizer_Tokenize(t *testing.T) {
	// soft-stub body was empty; add green SRX smokes
	tok := NewPolishSRXSentenceTokenizer()
	got := tok.Tokenize("To jest zdanie. To jest inne.")
	require.GreaterOrEqual(t, len(got), 2)
	require.Contains(t, got[0], "zdanie")
}
