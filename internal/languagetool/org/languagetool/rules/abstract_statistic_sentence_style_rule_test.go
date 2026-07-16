package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAbstractStatisticSentenceStyleRule(t *testing.T) {
	r := &AbstractStatisticSentenceStyleRule{
		MinPercent: 0,
		ConditionFulfilled: func(tokens []*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedTokenReadings {
			for _, t := range tokens {
				if t.GetToken() == "However" {
					return t
				}
			}
			return nil
		},
	}
	s1 := languagetool.AnalyzePlain("However this is long enough.")
	s2 := languagetool.AnalyzePlain("Short one.")
	matches := r.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.Len(t, matches, 1)
}
