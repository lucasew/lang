package tokenizers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSRXSentenceTokenizerPort(t *testing.T) {
	tok := NewSRXSentenceTokenizer("en")
	require.False(t, tok.SingleLineBreaksMarksPara())
	tok.SetSingleLineBreaksMarksParagraph(true)
	require.True(t, tok.SingleLineBreaksMarksPara())
	parts := tok.Tokenize("Hello world. Next sentence!")
	require.GreaterOrEqual(t, len(parts), 2)
	// re-join approx
	joined := ""
	for _, p := range parts {
		joined += p
	}
	require.Contains(t, joined, "Hello")
	require.Contains(t, joined, "Next")
}

func TestSimpleAsSentenceTokenizer(t *testing.T) {
	st := NewSimpleSentenceTokenizer().AsSentenceTokenizer()
	st.SetSingleLineBreaksMarksParagraph(true)
	require.True(t, st.SingleLineBreaksMarksPara())
	require.NotEmpty(t, st.Tokenize("A. B"))
}
