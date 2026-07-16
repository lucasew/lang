package disambiguation

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func sentenceStart() *languagetool.AnalyzedTokenReadings {
	// SENT_START pos like Java JLanguageTool.SENTENCE_START_TAGNAME
	tag := languagetool.SentenceStartTagName
	return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil))
}

func TestMultiWordChunkerSpace(t *testing.T) {
	c := NewMultiWordChunker([]string{"New York\tB-NP"}, MultiWordChunkerSettings{
		AllowFirstCapitalized: true,
	})
	// Include sentence-start so whitespace branch (j > 1) matches Java indexing.
	toks := []*languagetool.AnalyzedTokenReadings{
		sentenceStart(),
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
	require.True(t, found, "expected multiword POS tag on readings")
}

func TestMultiWordChunkerNoSpace(t *testing.T) {
	c := NewMultiWordChunker([]string{"...\tELLIPSIS"}, MultiWordChunkerSettings{})
	toks := []*languagetool.AnalyzedTokenReadings{
		sentenceStart(),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(".", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(".", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(".", nil, nil)),
	}
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	found := false
	for _, tok := range out.GetTokens() {
		for _, r := range tok.GetReadings() {
			if r.GetPOSTag() != nil {
				p := *r.GetPOSTag()
				if p == "ELLIPSIS" || p == "<ELLIPSIS>" || p == "</ELLIPSIS>" {
					found = true
				}
			}
		}
	}
	require.True(t, found)
}
