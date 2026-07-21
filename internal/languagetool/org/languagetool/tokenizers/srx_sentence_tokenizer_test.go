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

func TestSRXSentenceTokenizer_HonorsSrxPath(t *testing.T) {
	// Default segment.srx vs segment-simple.srx via SrxPath (Java constructor path).
	full := NewSRXSentenceTokenizer("en")
	simple := NewSRXSentenceTokenizerWithPath("xx", "/org/languagetool/tokenizers/segment-simple.srx")
	require.Equal(t, "/segment.srx", full.SrxPath)
	require.Equal(t, "/org/languagetool/tokenizers/segment-simple.srx", simple.SrxPath)

	// segment-simple Default breaks after ". " (no etc. exception).
	gotSimple := simple.Tokenize("Fruits, etc. Next sentence.")
	require.GreaterOrEqual(t, len(gotSimple), 2, "segment-simple Default breaks after '. '")
	require.Equal(t, "Fruits, etc. ", gotSimple[0])

	// createSrxDocument caches both resources
	doc1, err := cachedCreateSrxDocument("/segment.srx")
	require.NoError(t, err)
	require.NotNil(t, doc1)
	doc2, err := cachedCreateSrxDocument("/org/languagetool/tokenizers/segment-simple.srx")
	require.NoError(t, err)
	require.NotNil(t, doc2)
	require.NotEqual(t, doc1, doc2)
}
