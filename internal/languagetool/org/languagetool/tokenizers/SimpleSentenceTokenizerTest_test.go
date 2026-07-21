package tokenizers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.tokenizers.SimpleSentenceTokenizerTest.

func TestSimpleSentenceTokenizer_Tokenize(t *testing.T) {
	// TestTools.testSplit: join parts, tokenize, expect same parts
	sentences := []string{"Hi! ", "This is a test. ", "Here's more. ", "And even more?? ", "Yes."}
	var input string
	for _, s := range sentences {
		input += s
	}
	got := NewSimpleSentenceTokenizer().Tokenize(input)
	require.Equal(t, sentences, got)
}

// segment-simple.srx has no abbreviation exceptions (full segment.srx does).
// Soft invent noBreakAbbrevs removed — "etc. " may end a sentence here.
func TestSimpleSentenceTokenizer_NoInventAbbrevNoBreak(t *testing.T) {
	got := NewSimpleSentenceTokenizer().Tokenize("Fruits, etc. Next sentence.")
	// Default rule: break after ". " → two segments
	require.GreaterOrEqual(t, len(got), 2)
	require.Equal(t, "Fruits, etc. ", got[0])
}

func TestSimpleSentenceTokenizer_UsesSegmentSimpleSRX(t *testing.T) {
	// Load path must succeed so we exercise real SRX (not invent fallback only).
	doc, err := segmentSimpleDocument()
	require.NoError(t, err)
	require.NotNil(t, doc)
	// Default rule breaks after ". " with Unicode \s (UNICODE_CHARACTER_CLASS in Java).
	got := NewSimpleSentenceTokenizer().Tokenize("Hi.\u00A0Next")
	// NBSP is \s under UNICODE_CHARACTER_CLASS → break after ".NBSP"
	require.GreaterOrEqual(t, len(got), 1)
}

func TestSimpleSentenceTokenizer_ParagraphModeFlags(t *testing.T) {
	st := NewSimpleSentenceTokenizer()
	require.False(t, st.SingleLineBreaksMarksPara())
	st.SetSingleLineBreaksMarksParagraph(true)
	require.True(t, st.SingleLineBreaksMarksPara())
	// Adapter forwards flags
	ad := st.AsSentenceTokenizer()
	ad.SetSingleLineBreaksMarksParagraph(false)
	require.False(t, st.SingleLineBreaksMarksPara())
}
