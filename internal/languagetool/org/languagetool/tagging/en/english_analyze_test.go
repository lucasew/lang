package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of EnglishHybridDisambiguator multiword IGNORE_SPELLING for no-space "Qur'an".
func TestAnalyzeEnglishSentence_QuranIgnoreSpelling(t *testing.T) {
	sent := AnalyzeEnglishSentence("Qur'an")
	require.NotNil(t, sent)
	ignored := 0
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok == nil || tok.GetToken() == "" {
			continue
		}
		t.Logf("token %q ignored=%v", tok.GetToken(), tok.IsIgnoredBySpeller())
		if tok.IsIgnoredBySpeller() {
			ignored++
		}
	}
	require.GreaterOrEqual(t, ignored, 2, "expected multiword tokens ignore-spelling on Qur'an span")
}

func TestAnalyzeEnglishSentence_NewYorkPostIgnoreSpelling(t *testing.T) {
	// multiwords.txt: "New York Post\tNNP" (space multiword)
	sent := AnalyzeEnglishSentence("New York Post is big.")
	require.NotNil(t, sent)
	found := false
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "New" && tok.IsIgnoredBySpeller() {
			found = true
		}
	}
	require.True(t, found, "New in New York Post should ignore spelling")
}
