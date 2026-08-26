package rules

import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRemoteRuleFilters(t *testing.T) {
	f := NewRemoteRuleFilters()
	f.Register("en", &FilterRule{
		IDPattern: regexp.MustCompile(`^AI_.*`),
		MatchPositions: func(s *languagetool.AnalyzedSentence) []MatchPosition {
			return []MatchPosition{{Start: 0, End: 4}}
		},
	})
	sent := languagetool.AnalyzePlain("test sentence")
	keep := NewRuleMatch(NewFakeRule("MORFO"), sent, 0, 4, "x")
	drop := NewRuleMatch(NewFakeRule("AI_FOO"), sent, 0, 4, "y")
	other := NewRuleMatch(NewFakeRule("AI_BAR"), sent, 5, 8, "z")

	out := f.FilterMatches("en", sent, []*RuleMatch{keep, drop, other})
	require.Len(t, out, 2)
	require.Equal(t, "MORFO", ruleIDOfMatch(out[0]))
	require.Equal(t, "AI_BAR", ruleIDOfMatch(out[1]))

	require.Equal(t, "en/remote-rule-filters.xml", GetRemoteRuleFilterFilename("en-US"))
}
