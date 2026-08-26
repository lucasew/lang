package pt

import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestPortugueseRemoteRuleFilter_Rules(t *testing.T) {
	f := rules.NewRemoteRuleFilters()
	f.Register("pt", &rules.FilterRule{
		IDPattern: regexp.MustCompile(`AI_.*`),
		MatchPositions: func(s *languagetool.AnalyzedSentence) []rules.MatchPosition {
			return []rules.MatchPosition{{Start: 0, End: 2}}
		},
	})
	sent := languagetool.AnalyzePlain("ab cd")
	drop := rules.NewRuleMatch(rules.NewFakeRule("AI_X"), sent, 0, 2, "d")
	keep := rules.NewRuleMatch(rules.NewFakeRule("OTHER"), sent, 0, 2, "k")
	out := f.FilterMatches("pt", sent, []*rules.RuleMatch{drop, keep})
	require.Len(t, out, 1)
}
