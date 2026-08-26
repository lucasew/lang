package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAbstractStatisticStyleRule(t *testing.T) {
	r := &AbstractStatisticStyleRule{
		MinPercent: 0,
		ConditionFulfilled: func(tokens []*languagetool.AnalyzedTokenReadings, i int) int {
			if tokens[i].GetToken() == "very" {
				return i
			}
			return -1
		},
	}
	sent := languagetool.AnalyzePlain("a very good very day")
	matches := r.MatchList([]*languagetool.AnalyzedSentence{sent})
	require.Len(t, matches, 2)
}
