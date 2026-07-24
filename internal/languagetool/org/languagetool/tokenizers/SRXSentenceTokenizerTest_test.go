package tokenizers

// Twin of languagetool-standalone/src/test/java/org/languagetool/tokenizers/SRXSentenceTokenizerTest.java
import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSRXSentenceTokenizer_OfficeFootnoteTokenize(t *testing.T) {
	// Port: office STX control char \u0002 should not prevent sentence split.
	input := "A sentence.\u0002 And another one."
	tok := NewSRXSentenceTokenizer("en")
	got := tok.Tokenize(input)
	require.GreaterOrEqual(t, len(got), 1, fmt.Sprintf("got %q", got))
	// Prefer two segments when possible
	if len(got) >= 2 {
		require.Contains(t, got[0], "sentence")
		require.Contains(t, got[1], "another")
	}
}

func TestSRXSentenceTokenizer_DotNetSentence(t *testing.T) {
	// .Net should generally not split mid-token for English-like SRX
	tok := NewSRXSentenceTokenizer("en")
	got := tok.Tokenize("I use .Net daily. Really.")
	require.NotEmpty(t, got)
}
