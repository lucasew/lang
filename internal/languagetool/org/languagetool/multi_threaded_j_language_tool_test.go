package languagetool

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultiThreadedJLanguageTool(t *testing.T) {
	lt := NewMultiThreadedJLanguageTool("en", 2)
	require.Equal(t, 2, lt.GetThreadPoolSize())
	var n atomic.Int32
	lt.Matchers = []SentenceMatcherFunc{
		func(s *AnalyzedSentence) error {
			n.Add(1)
			return nil
		},
	}
	sents := []*AnalyzedSentence{
		NewAnalyzedSentence([]*AnalyzedTokenReadings{NewAnalyzedTokenReadings(NewAnalyzedToken("a", nil, nil))}),
		NewAnalyzedSentence([]*AnalyzedTokenReadings{NewAnalyzedTokenReadings(NewAnalyzedToken("b", nil, nil))}),
		NewAnalyzedSentence([]*AnalyzedTokenReadings{NewAnalyzedTokenReadings(NewAnalyzedToken("c", nil, nil))}),
	}
	require.NoError(t, lt.CheckSentences(sents))
	require.Equal(t, int32(3), n.Load())
	lt.Shutdown()
}

func TestMultiThreadedJLanguageTool_ParallelCheck(t *testing.T) {
	lt := NewMultiThreadedJLanguageTool("en", 4)
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
	// multi-sentence document
	src := "This is is one. And an test two. Fine here."
	m := lt.Check(src)
	require.NotEmpty(t, m)
	// both rule types may fire
	ids := map[string]bool{}
	for _, x := range m {
		ids[x.RuleID] = true
		require.GreaterOrEqual(t, x.FromPos, 0)
		require.LessOrEqual(t, x.ToPos, len(src))
	}
	require.True(t, ids["WORD_REPEAT_RULE"] || ids["EN_A_VS_AN"])
	lt.Shutdown()
	require.True(t, lt.IsShutdown())
}
