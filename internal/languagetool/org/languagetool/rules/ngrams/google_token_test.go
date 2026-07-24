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

// Twin of GoogleToken.getGoogleTokens: token.length() is UTF-16, not UTF-8 bytes.
// "café " = c a f é space → UTF-16 len 5; "ok" starts at 5 (not 6).
func TestGetGoogleTokens_UTF16Positions(t *testing.T) {
	// Fixed tokenizer: surface tokens as given (whitespace included).
	tok := func(s string) []string {
		// "café ok" → ["café", " ", "ok"] for this test only
		if s == "café ok" {
			return []string{"café", " ", "ok"}
		}
		if s == "😀 x" {
			// emoji is one rune / two UTF-16 units
			return []string{"😀", " ", "x"}
		}
		return []string{s}
	}
	tokens := GetGoogleTokens("café ok", false, tok)
	require.Len(t, tokens, 2)
	require.Equal(t, "café", tokens[0].Token)
	require.Equal(t, 0, tokens[0].StartPos)
	require.Equal(t, 4, tokens[0].EndPos) // c,a,f,é each 1 UTF-16 unit
	require.Equal(t, "ok", tokens[1].Token)
	require.Equal(t, 5, tokens[1].StartPos) // after "café "
	require.Equal(t, 7, tokens[1].EndPos)

	emoji := GetGoogleTokens("😀 x", false, tok)
	require.Len(t, emoji, 2)
	require.Equal(t, "😀", emoji[0].Token)
	require.Equal(t, 0, emoji[0].StartPos)
	require.Equal(t, 2, emoji[0].EndPos) // surrogate pair
	require.Equal(t, "x", emoji[1].Token)
	require.Equal(t, 3, emoji[1].StartPos) // after "😀 "
	require.Equal(t, 4, emoji[1].EndPos)
}
