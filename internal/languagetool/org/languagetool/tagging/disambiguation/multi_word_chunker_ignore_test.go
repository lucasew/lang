package disambiguation

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestMultiWordChunker_IgnoreSpelling(t *testing.T) {
	c := NewMultiWordChunker([]string{"New York\tB-NP"}, MultiWordChunkerSettings{
		AllowFirstCapitalized: true,
	})
	c.SetIgnoreSpelling(true)
	toks := []*languagetool.AnalyzedTokenReadings{
		sentenceStart(),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("New", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("York", nil, nil)),
	}
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	require.NotNil(t, out)
	// New and York should be ignored by speller (range from multiword match)
	ignored := 0
	for _, tok := range out.GetTokens() {
		if tok != nil && tok.IsIgnoredBySpeller() {
			ignored++
		}
	}
	require.GreaterOrEqual(t, ignored, 2, "expected multiword tokens ignore-spelling")
}

func TestMultiWordChunker_SetIgnoreSpelling(t *testing.T) {
	c := NewMultiWordChunker([]string{"ab\tX"}, MultiWordChunkerSettings{})
	c.SetIgnoreSpelling(true)
	require.True(t, c.AddIgnoreSpelling)
}
