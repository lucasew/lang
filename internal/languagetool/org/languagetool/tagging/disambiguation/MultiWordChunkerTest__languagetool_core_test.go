package disambiguation

// Twin of MultiWordChunkerTest (core)
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func hasPOS(sent *languagetool.AnalyzedSentence, want string) bool {
	for _, tok := range sent.GetTokens() {
		for _, r := range tok.GetReadings() {
			if r.GetPOSTag() != nil && *r.GetPOSTag() == want {
				return true
			}
		}
	}
	return false
}

func TestMultiWordChunker_languagetool_core_Disambiguate1(t *testing.T) {
	c := NewMultiWordChunker([]string{"New York\tB-NP"}, MultiWordChunkerSettings{AllowFirstCapitalized: true})
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("New", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("York", nil, nil)),
	}
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	require.True(t, hasPOS(out, "B-NP") || hasPOS(out, "<B-NP>") || hasPOS(out, "</B-NP>"))
	// non-match sentence
	toks2 := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Los", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Angeles", nil, nil)),
	}
	out2 := c.Disambiguate(languagetool.NewAnalyzedSentence(toks2))
	require.False(t, hasPOS(out2, "B-NP") || hasPOS(out2, "<B-NP>"))
}

func TestMultiWordChunker_languagetool_core_Disambiguate2(t *testing.T) {
	c := NewMultiWordChunker2([]string{"New York\tB-NP"}, true)
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("New", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("York", nil, nil)),
	}
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	require.True(t, hasPOS(out, "<B-NP>") || hasPOS(out, "B-NP"))
}

func TestMultiWordChunker_languagetool_core_Disambiguate2NoMatch(t *testing.T) {
	c := NewMultiWordChunker2([]string{"New York\tB-NP"}, true)
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Old", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("York", nil, nil)),
	}
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	require.False(t, hasPOS(out, "<B-NP>"))
}

func TestMultiWordChunker_languagetool_core_Disambiguate2RemoveOtherReadings(t *testing.T) {
	c := NewMultiWordChunker2([]string{"New York\tB-NP"}, true)
	c.SetRemoveOtherReadings(true)
	c.SetWrapTag(true)
	tag := languagetool.SentenceStartTagName
	nn := "NN"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("New", &nn, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("York", &nn, nil)),
	}
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	require.True(t, hasPOS(out, "<B-NP>") || hasPOS(out, "B-NP"))
}
