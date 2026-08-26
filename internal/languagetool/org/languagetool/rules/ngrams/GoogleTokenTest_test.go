package ngrams

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
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

// Port of GoogleTokenTest.testTokenizationWithPosTag — align POS from AnalyzedSentence.
func TestGoogleToken_TokenizationWithPosTag(t *testing.T) {
	// Build sentence with known positions matching WordTokenizer UTF-16 layout.
	text := "Hello world"
	sent := languagetool.AnalyzePlain(text)
	// inject POS on non-start tokens
	tokens := sent.GetTokens()
	require.GreaterOrEqual(t, len(tokens), 3)
	// find "Hello" and "world"
	for _, tr := range tokens {
		if tr.GetToken() == "Hello" {
			p, l := "NNP", "Hello"
			tr.AddReading(languagetool.NewAnalyzedToken("Hello", &p, &l), "test")
		}
		if tr.GetToken() == "world" {
			p, l := "NN", "world"
			tr.AddReading(languagetool.NewAnalyzedToken("world", &p, &l), "test")
		}
	}
	wt := tokenizers.NewWordTokenizer()
	gts := GetGoogleTokensFromSentence(sent, true, wt.Tokenize)
	require.Equal(t, GoogleSentenceStart, gts[0].Token)
	// find Hello google token with POS
	var foundHello bool
	for _, g := range gts {
		if g.Token == "Hello" {
			foundHello = true
			require.NotEmpty(t, g.PosTags, "POS should align from AnalyzedSentence")
			require.NotNil(t, g.PosTags[0].GetPOSTag())
			require.Equal(t, "NNP", *g.PosTags[0].GetPOSTag())
		}
	}
	require.True(t, foundHello)
}

// findOriginalAnalyzedTokens: only exact UTF-16 start/end on tokensWithoutWhitespace (Java twin).
func TestGoogleToken_FindOriginal_UTF16Align(t *testing.T) {
	// "café ok" — AnalyzePlain sets UTF-16 start positions for tokens.
	sent := languagetool.AnalyzePlain("café ok")
	// Inject POS on "ok" so alignment is observable.
	for _, tr := range sent.GetTokensWithoutWhitespace() {
		if tr.GetToken() == "ok" {
			p, l := "JJ", "ok"
			tr.AddReading(languagetool.NewAnalyzedToken("ok", &p, &l), "test")
		}
	}
	// Tokenizer that yields the same surfaces/spaces as AnalyzePlain expects.
	tok := func(s string) []string {
		return []string{"café", " ", "ok"}
	}
	gts := GetGoogleTokensFromSentence(sent, false, tok)
	require.Len(t, gts, 2)
	require.Equal(t, "café", gts[0].Token)
	require.Equal(t, 0, gts[0].StartPos)
	require.Equal(t, 4, gts[0].EndPos)
	require.Equal(t, "ok", gts[1].Token)
	// "café " is 5 UTF-16 units → "ok" at 5..7
	require.Equal(t, 5, gts[1].StartPos)
	require.Equal(t, 7, gts[1].EndPos)
	// POS aligns only when LT token span matches UTF-16 indices (exact start/end).
	require.NotEmpty(t, gts[1].PosTags)
	var sawJJ bool
	for _, at := range gts[1].PosTags {
		if at != nil && at.GetPOSTag() != nil && *at.GetPOSTag() == "JJ" {
			sawJJ = true
			break
		}
	}
	require.True(t, sawJJ, "injected JJ reading must align via UTF-16 span")
}
