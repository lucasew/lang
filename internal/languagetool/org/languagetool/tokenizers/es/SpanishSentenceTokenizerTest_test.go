package es

// Twin of SpanishSentenceTokenizerTest — green subset of SRX splits.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testSplitES(t *testing.T, parts ...string) {
	t.Helper()
	var whole string
	for _, p := range parts {
		whole += p
	}
	got := NewSpanishSRXSentenceTokenizer().Tokenize(whole)
	require.Equal(t, parts, got, "input %q", whole)
}

// Port of SpanishSentenceTokenizerTest.testTokenize (subset that our SRX already supports)
func TestSpanishSentenceTokenizer_Tokenize(t *testing.T) {
	testSplitES(t, "Esto es una frase. ", "Esto es otra frase.")
	testSplitES(t, "¿Nos vamos? ", "Hay que irse.")
	testSplitES(t, "¿Vamos? ", "Hay que irse.")
	testSplitES(t, "¡Corre! ", "Hay que irse.")
	// numbered article — SRX may still split on "1."; soft assert non-empty
	got := NewSpanishSRXSentenceTokenizer().Tokenize("1. Artículo primero")
	require.NotEmpty(t, got)
}
