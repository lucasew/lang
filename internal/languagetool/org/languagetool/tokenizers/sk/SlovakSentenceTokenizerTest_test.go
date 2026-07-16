package sk

// Twin of SlovakSentenceTokenizerTest — Java has no @Test; SRX green smoke.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlovakSentenceTokenizer_NoTests(t *testing.T) {
	got := NewSlovakSRXSentenceTokenizer().Tokenize("Ahoj. Ako sa máš?")
	require.GreaterOrEqual(t, len(got), 2)
}
