package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuppressIfAnyRuleMatchesFilter(t *testing.T) {
	f := NewSuppressIfAnyRuleMatchesFilter(func(ruleID, sent string) []MatchSpan {
		if ruleID == "BAD" && strings.Contains(sent, "oops") {
			return []MatchSpan{{From: 0, To: 4}}
		}
		return nil
	})
	require.True(t, f.ShouldSuppress("hi X there", 3, 4, []string{"oops"}, "BAD"))
	require.False(t, f.ShouldSuppress("hi X there", 3, 4, []string{"fine"}, "BAD"))
}
