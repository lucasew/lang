package ngrams

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

func TestGetGoogleTokens(t *testing.T) {
	wt := tokenizers.NewWordTokenizer()
	tok := func(s string) []string { return wt.Tokenize(s) }
	tokens := GetGoogleTokens("Hello world", true, tok)
	require.Equal(t, GoogleSentenceStart, tokens[0].Token)
	// non-whitespace only
	var words []string
	for _, g := range tokens {
		if g.Token != GoogleSentenceStart {
			words = append(words, g.Token)
		}
	}
	require.Equal(t, []string{"Hello", "world"}, words)

	// apostrophe normalize
	g := NewGoogleToken("’", 0, 1)
	require.Equal(t, "'", g.Token)

	strs := GetGoogleTokensForString("a b", false, tok)
	require.Equal(t, []string{"a", "b"}, strs)
	_ = strings.Join // silence
}
