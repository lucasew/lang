package spelling

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCanBeIgnoredToken_UrlAndEmail(t *testing.T) {
	require.True(t, IsUrl("https://www.languagetool.org"))
	require.True(t, IsEMail("martin.mustermann@test.de"))
	// token-level
	sent := languagetool.AnalyzePlain("hello")
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "hello" {
			require.False(t, CanBeIgnoredToken(tok))
		}
	}
	// immunized token
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "hello" {
			tok.Immunize(0)
			require.True(t, CanBeIgnoredToken(tok))
		}
	}
}

func TestAcceptWord_IgnoreWordsWithLength(t *testing.T) {
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_EN_US", "spell", "en")
	r.IsMisspelled = func(string) bool { return true }
	r.IgnoreWordsWithLength = 1
	require.True(t, r.AcceptWord("a"))
	require.True(t, r.AcceptWord("I"))
	require.False(t, r.AcceptWord("ab"))
	// prohibited still wins
	r.AddProhibitedWords("a")
	require.False(t, r.AcceptWord("a"))
}

func TestCanBeIgnoredToken_UrlSurface(t *testing.T) {
	// Build a single-token sentence if tokenizer keeps URL whole
	toks := languagetool.WordTokenizerForLanguage("en").Tokenize("https://example.com/foo")
	// Find a URL-shaped token
	foundURL := false
	for _, s := range toks {
		if IsUrl(s) {
			foundURL = true
			atr := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(s, nil, nil))
			require.True(t, CanBeIgnoredToken(atr))
		}
	}
	if !foundURL {
		// still verify helper
		require.True(t, IsUrl("https://example.com/foo"))
	}
}
