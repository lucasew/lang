package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLanguageDependentRuleMatchFilter(t *testing.T) {
	f := NewLanguageDependentRuleMatchFilter(func(matches []*RuleMatch, text string, enabled map[string]struct{}) []*RuleMatch {
		var out []*RuleMatch
		for _, m := range matches {
			if r, ok := m.Rule.(interface{ GetID() string }); ok {
				if _, ok := enabled[r.GetID()]; ok {
					out = append(out, m)
				}
			}
		}
		return out
	}, []string{"KEEP"})
	m1 := NewRuleMatch(NewFakeRule("KEEP"), nil, 0, 1, "a")
	m2 := NewRuleMatch(NewFakeRule("DROP"), nil, 0, 1, "b")
	out := f.Apply([]*RuleMatch{m1, m2}, "text")
	require.Len(t, out, 1)
	require.Equal(t, "KEEP", out[0].Rule.(interface{ GetID() string }).GetID())
}
