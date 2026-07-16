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
