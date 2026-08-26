package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnalyzeWithTagger(t *testing.T) {
	s := AnalyzeWithTagger("cats run", func(tok string) []TokenTag {
		switch tok {
		case "cats":
			return []TokenTag{{POS: "NNS", Lemma: "cat"}}
		case "run":
			return []TokenTag{{POS: "VBP", Lemma: "run"}}
		}
		return nil
	})
	toks := s.GetTokensWithoutWhitespace()
	// skip SENT_START
	require.GreaterOrEqual(t, len(toks), 2)
	// first non-start
	var cat *AnalyzedTokenReadings
	for _, t := range toks {
		if t.GetToken() == "cats" {
			cat = t
			break
		}
	}
	require.NotNil(t, cat)
	require.Equal(t, "NNS", *cat.GetReadings()[0].GetPOSTag())
	require.Equal(t, "cat", *cat.GetReadings()[0].GetLemma())
}

func TestJLanguageTool_TagWordInject(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.TagWord = func(tok string) []TokenTag {
		if tok == "dogs" {
			return []TokenTag{{POS: "NNS", Lemma: "dog"}}
		}
		return nil
	}
	sents := lt.Analyze("dogs bark.")
	require.NotEmpty(t, sents)
	found := false
	for _, s := range sents {
		for _, tok := range s.GetTokensWithoutWhitespace() {
			if tok.GetToken() == "dogs" {
				pos := tok.GetReadings()[0].GetPOSTag()
				require.NotNil(t, pos)
				require.Equal(t, "NNS", *pos)
				found = true
			}
		}
	}
	require.True(t, found)
}
