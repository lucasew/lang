package tokenizers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.tokenizers.SimpleSentenceTokenizerTest.

// TestSimpleSentenceTokenizer_Tokenize ports testTokenize → private testSplit
// → TestTools.testSplit(sentences, tokenizer): join parts, tokenize, equal parts.
func TestSimpleSentenceTokenizer_Tokenize(t *testing.T) {
	testSplitSimple(t, "Hi! ", "This is a test. ", "Here's more. ", "And even more?? ", "Yes.")
}

// testSplitSimple mirrors Java SimpleSentenceTokenizerTest.testSplit + TestTools.testSplit.
func testSplitSimple(t *testing.T, sentences ...string) {
	t.Helper()
	tokenizer := NewSimpleSentenceTokenizer()
	var input string
	for _, s := range sentences {
		input += s
	}
	require.Equal(t, []string(sentences), tokenizer.Tokenize(input))
}

// segment-simple.srx has no abbreviation exceptions (full segment.srx does).
// Official Default rule breaks after ". " — exact segments, not soft len checks.
func TestSimpleSentenceTokenizer_NoInventAbbrevNoBreak(t *testing.T) {
	got := NewSimpleSentenceTokenizer().Tokenize("Fruits, etc. Next sentence.")
	require.Equal(t, []string{"Fruits, etc. ", "Next sentence."}, got)
}

// Construction path: language "xx", SRX path segment-simple.srx; paragraph flags
// from SRXSentenceTokenizer (Java inheritance).
func TestSimpleSentenceTokenizer_ConstructionAndParagraphFlags(t *testing.T) {
	st := NewSimpleSentenceTokenizer()
	require.Equal(t, "xx", st.LanguageCode)
	require.Equal(t, "/org/languagetool/tokenizers/segment-simple.srx", st.SrxPath)

	// Java constructor: setSingleLineBreaksMarksParagraph(false) → parCode "_two"
	require.False(t, st.SingleLineBreaksMarksPara())
	st.SetSingleLineBreaksMarksParagraph(true)
	require.True(t, st.SingleLineBreaksMarksPara())
	st.SetSingleLineBreaksMarksParagraph(false)
	require.False(t, st.SingleLineBreaksMarksPara())

	// is-a SentenceTokenizer (Java extends)
	var as SentenceTokenizer = st
	as.SetSingleLineBreaksMarksParagraph(true)
	require.True(t, st.SingleLineBreaksMarksPara())
	require.True(t, as.SingleLineBreaksMarksPara())

	// Official resource loads (SrxTools.createSrxDocument twin)
	doc, err := segmentSimpleDocument()
	require.NoError(t, err)
	require.NotNil(t, doc)
}
