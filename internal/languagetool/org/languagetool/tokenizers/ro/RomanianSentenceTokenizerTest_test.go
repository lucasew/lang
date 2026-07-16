package ro

// Twin of RomanianSentenceTokenizerTest — Java has no @Test; SRX green smoke.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRomanianSentenceTokenizer_NoTests(t *testing.T) {
	got := NewRomanianSRXSentenceTokenizer().Tokenize("Salut. Cum ești?")
	require.GreaterOrEqual(t, len(got), 2)
}
