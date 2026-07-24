package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestIsEnglishWordFilter(t *testing.T) {
	nn := "NN"
	tag := func(word string) *languagetool.AnalyzedTokenReadings {
		if word == "cat" {
			return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(word, &nn, nil))
		}
		// untagged
		return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(word, nil, nil))
	}
	f := NewIsEnglishWordFilter(tag)
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("cat", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("xyzzy", nil, nil)),
	}
	m := NewRuleMatch(NewFakeRule("R"), nil, 0, 3, "msg")
	// form 1 = cat → tagged
	require.NotNil(t, f.AcceptRuleMatch(m, map[string]string{"formPositions": "1"}, 0, toks, []int{1, 1}))
	// form 2 = xyzzy → not tagged
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"formPositions": "2"}, 0, toks, []int{1, 1}))
	// with postag
	require.NotNil(t, f.AcceptRuleMatch(m, map[string]string{"formPositions": "1", "postags": "NN"}, 0, toks, []int{1, 1}))
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"formPositions": "1", "postags": "VB"}, 0, toks, []int{1, 1}))
	// no tagger
	require.Nil(t, NewIsEnglishWordFilter(nil).AcceptRuleMatch(m, map[string]string{"formPositions": "1"}, 0, toks, []int{1}))
}
