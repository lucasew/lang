package disambiguation

// Twin of PolishDisambiguationRuleTest.testChunker
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPolishDisambiguationRule_Chunker(t *testing.T) {
	// Multi-word chunker surface used by language disambiguators.
	c := NewMultiWordChunker([]string{"New York\tB-NP"}, MultiWordChunkerSettings{AllowFirstCapitalized: true})
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("New", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("York", nil, nil)),
	}
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	require.NotNil(t, out)
	found := false
	for _, tok := range out.GetTokens() {
		for _, r := range tok.GetReadings() {
			if r.GetPOSTag() != nil {
				p := *r.GetPOSTag()
				if p == "B-NP" || p == "<B-NP>" || p == "</B-NP>" {
					found = true
				}
			}
		}
	}
	require.True(t, found)
}
