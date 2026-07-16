package ca

// Twin of CatalanSentenceTokenizerTest — Java has no @Test; SRX green smoke.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatalanSentenceTokenizer_NoTests(t *testing.T) {
	got := NewCatalanSRXSentenceTokenizer().Tokenize("Hola. Com estàs?")
	require.GreaterOrEqual(t, len(got), 2)
}
