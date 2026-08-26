package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.RuleWithMaxFilterTest.

func TestRuleWithMaxFilter_Filter(t *testing.T) {
	rule1 := NewPatternRule("id1")
	rule2 := NewPatternRule("id1")
	match1 := NewRuleMatch(rule1, nil, 10, 20, "Match1")
	match2 := NewRuleMatch(rule2, nil, 15, 25, "Match2")
	filter := NewRuleWithMaxFilter()
	require.Len(t, filter.Filter([]*RuleMatch{match1, match2}), 2)

	match3 := NewRuleMatch(rule2, nil, 11, 19, "Match3")
	require.Len(t, filter.Filter([]*RuleMatch{match1, match3}), 1)
}

func TestRuleWithMaxFilter_NoFilteringIfNotOverlapping(t *testing.T) {
	rule1 := NewPatternRule("id1")
	rule2 := NewPatternRule("id1")
	match1 := NewRuleMatch(rule1, nil, 10, 20, "Match1")
	match2 := NewRuleMatch(rule2, nil, 21, 25, "Match2")
	require.Len(t, NewRuleWithMaxFilter().Filter([]*RuleMatch{match1, match2}), 2)
}

func TestRuleWithMaxFilter_NoFilteringIfDifferentRulegroups(t *testing.T) {
	rule1 := NewPatternRule("id1")
	rule2 := NewPatternRule("id2")
	match1 := NewRuleMatch(rule1, nil, 10, 20, "Match1")
	match2 := NewRuleMatch(rule2, nil, 15, 25, "Match2")
	filter := NewRuleWithMaxFilter()
	require.Len(t, filter.Filter([]*RuleMatch{match1, match2}), 2)
	match3 := NewRuleMatch(rule2, nil, 11, 19, "Match3")
	require.Len(t, filter.Filter([]*RuleMatch{match1, match3}), 2)
}

func TestRuleWithMaxFilter_Overlaps(t *testing.T) {
	filter := NewRuleWithMaxFilter()
	require.True(t, filter.Includes(makeMaxRuleMatch(10, 20), makeMaxRuleMatch(10, 20)))
	require.False(t, filter.Includes(makeMaxRuleMatch(10, 20), makeMaxRuleMatch(5, 11)))
	require.False(t, filter.Includes(makeMaxRuleMatch(10, 20), makeMaxRuleMatch(11, 21)))
	require.True(t, filter.Includes(makeMaxRuleMatch(10, 20), makeMaxRuleMatch(11, 19)))
	require.False(t, filter.Includes(makeMaxRuleMatch(10, 20), makeMaxRuleMatch(1, 10)))
	require.True(t, filter.Includes(makeMaxRuleMatch(10, 20), makeMaxRuleMatch(19, 20)))

	require.False(t, filter.Includes(makeMaxRuleMatch(10, 20), makeMaxRuleMatch(21, 30)))
	require.False(t, filter.Includes(makeMaxRuleMatch(10, 20), makeMaxRuleMatch(1, 9)))
}

func makeMaxRuleMatch(fromPos, toPos int) *RuleMatch {
	return NewRuleMatch(NewFakeRule(""), nil, fromPos, toPos, "FakeMatch1")
}
