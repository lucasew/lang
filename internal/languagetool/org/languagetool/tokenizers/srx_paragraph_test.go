package tokenizers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSRX_PreservesParagraphBreaks(t *testing.T) {
	tok := NewSRXSentenceTokenizer("en")
	// Java: "He won't\n\n", "Really."
	got := tok.Tokenize("He won't\n\nReally.")
	require.Len(t, got, 2)
	require.Equal(t, "He won't\n\n", got[0])
	require.Equal(t, "Really.", got[1])

	// four newlines → empty line mark on first sentence
	got = tok.Tokenize("Hello world.\n\n\n\nNext para.")
	require.Len(t, got, 2)
	require.Equal(t, "Hello world.\n\n\n\n", got[0])
	require.Equal(t, "Next para.", got[1])
}
