package fr

// Twin of FrenchSentenceTokenizerTest — SRX green smokes.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFrenchSentenceTokenizer_NoTests(t *testing.T) {
	// Java class has no @Test methods; exercise tokenizer surface anyway.
	got := NewFrenchSRXSentenceTokenizer().Tokenize("Bonjour. Comment allez-vous?")
	require.Len(t, got, 2)
	require.Equal(t, "Bonjour. ", got[0])
	require.Equal(t, "Comment allez-vous?", got[1])
}
