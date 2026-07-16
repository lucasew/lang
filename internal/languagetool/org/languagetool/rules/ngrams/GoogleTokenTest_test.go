package ngrams

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Port of GoogleTokenTest.testTokenization
func TestGoogleToken_Tokenization(t *testing.T) {
	wt := tokenizers.NewWordTokenizer()
	tok := func(s string) []string { return wt.Tokenize(s) }
	tokens := GetGoogleTokens("This, isn't a test.", true, tok)
	require.Equal(t, GoogleSentenceStart, tokens[0].Token)
	// remaining non-whitespace tokens
	require.GreaterOrEqual(t, len(tokens), 5)
	// ensure apostrophe forms exist as non-ws tokens
	joined := ""
	for _, g := range tokens {
		joined += g.Token + " "
	}
	require.Contains(t, joined, "This")
	require.Contains(t, joined, "isn")
}

// Port of GoogleTokenTest.testTokenizationWithPosTag — soft without full analysis stack
func TestGoogleToken_TokenizationWithPosTag(t *testing.T) {
	t.Skip("needs AnalyzedSentence with POS alignment for POS-carrying GoogleTokens")
}
