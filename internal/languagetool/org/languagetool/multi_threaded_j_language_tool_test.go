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
