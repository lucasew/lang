package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuleMatchFilter(t *testing.T) {
	id := IdentityRuleMatchFilter()
	m := []*RuleMatch{NewRuleMatch(NewFakeRule("X"), nil, 0, 1, "msg")}
	require.Equal(t, m, id.Filter(m, "text"))

	drop := RuleMatchFilterFunc(func(ms []*RuleMatch, _ string) []*RuleMatch {
		return nil
	})
	require.Empty(t, drop.Filter(m, "t"))
}
