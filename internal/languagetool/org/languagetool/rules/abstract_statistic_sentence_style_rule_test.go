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

// Twin of Java AbstractStatisticSentenceStyleRule MARKS_REGEX range quirk (!-–).
func TestStatisticMarksRE_JavaRangeQuirk(t *testing.T) {
	require.True(t, statMarksRE.MatchString(","))
	require.True(t, statMarksRE.MatchString("."))
	require.True(t, statMarksRE.MatchString("!"))
	require.True(t, statMarksRE.MatchString("a"))
	require.True(t, statMarksRE.MatchString("5"))
	require.True(t, statMarksRE.MatchString("—"))
	require.True(t, statMarksRE.MatchString("•"))
	require.False(t, statMarksRE.MatchString("Auto"))
	require.False(t, statMarksRE.MatchString("ab"))
	// isMark helper uses the same RE
	tok := languagetool.AnalyzePlain("a").GetTokensWithoutWhitespace()
	require.NotEmpty(t, tok)
	// Find token "a"
	var aTok *languagetool.AnalyzedTokenReadings
	for _, t := range tok {
		if t != nil && t.GetToken() == "a" {
			aTok = t
			break
		}
	}
	require.NotNil(t, aTok)
	require.True(t, IsStatisticMark(aTok))
}
