package disambiguation

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestMultiWordChunker2(t *testing.T) {
	c := NewMultiWordChunker2([]string{"New York\tB-NP"}, true)
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("New", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("York", nil, nil)),
	}
	// Snapshot original readings count so we can prove input is not mutated.
	origNewReadings := len(toks[1].GetReadings())
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	found := false
	for _, tok := range out.GetTokens() {
		for _, r := range tok.GetReadings() {
			if r.GetPOSTag() != nil && *r.GetPOSTag() == "<B-NP>" {
				found = true
			}
		}
	}
	require.True(t, found)
	// Java setAndAnnotate builds a new ATR — input token readings stay unchanged.
	require.Equal(t, origNewReadings, len(toks[1].GetReadings()), "must not mutate input tokens")
	// Output "New" must be a different ATR with the multiword reading attached.
	outNew := out.GetTokens()[1]
	require.NotSame(t, toks[1], outNew)
	require.Greater(t, len(outNew.GetReadings()), origNewReadings)
}

func TestMultiWordChunker2_RemoveOtherReadings(t *testing.T) {
	c := NewMultiWordChunker2([]string{"foo bar\tMW"}, false)
	c.SetRemoveOtherReadings(true)
	c.SetWrapTag(false)
	nn := "NN"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("foo", &nn, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("bar", &nn, nil)),
	}
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	foo := out.GetTokens()[1]
	require.Len(t, foo.GetReadings(), 1)
	require.Equal(t, "MW", *foo.GetReadings()[0].GetPOSTag())
}
