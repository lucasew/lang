package disambiguation

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestMultiWordChunker2_IgnoreSpelling(t *testing.T) {
	c := NewMultiWordChunker2([]string{"New York\tB-NP"}, true)
	c.SetIgnoreSpelling(true)
	toks := []*languagetool.AnalyzedTokenReadings{
		sentenceStart(),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("New", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("York", nil, nil)),
	}
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	require.NotNil(t, out)
	ignored := 0
	for _, tok := range out.GetTokens() {
		if tok != nil && tok.IsIgnoredBySpeller() {
			ignored++
		}
	}
	require.GreaterOrEqual(t, ignored, 2)
}
