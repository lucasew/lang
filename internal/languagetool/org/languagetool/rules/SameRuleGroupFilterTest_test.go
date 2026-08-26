package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.SameRuleGroupFilterTest.

func TestSameRuleGroupFilter_Filter(t *testing.T) {
	rule1 := NewPatternRule("id1")
	rule2 := NewPatternRule("id1")
	match1 := NewRuleMatch(rule1, nil, 10, 20, "Match1")
	match2 := NewRuleMatch(rule2, nil, 15, 25, "Match2")
	filtered := NewSameRuleGroupFilter().Filter([]*RuleMatch{match1, match2})
	require.Len(t, filtered, 1)
	require.Equal(t, "Match1", filtered[0].GetMessage())
}

func TestSameRuleGroupFilter_NoFilteringIfNotOverlapping(t *testing.T) {
	rule1 := NewPatternRule("id1")
	rule2 := NewPatternRule("id1")
	match1 := NewRuleMatch(rule1, nil, 10, 20, "Match1")
	match2 := NewRuleMatch(rule2, nil, 21, 25, "Match2")
	filtered := NewSameRuleGroupFilter().Filter([]*RuleMatch{match1, match2})
	require.Len(t, filtered, 2)
}

func TestSameRuleGroupFilter_NoFilteringIfDifferentRulegroups(t *testing.T) {
	rule1 := NewPatternRule("id1")
	rule2 := NewPatternRule("id2")
	match1 := NewRuleMatch(rule1, nil, 10, 20, "Match1")
	match2 := NewRuleMatch(rule2, nil, 15, 25, "Match2")
	filtered := NewSameRuleGroupFilter().Filter([]*RuleMatch{match1, match2})
	require.Len(t, filtered, 2)
}

func TestSameRuleGroupFilter_Overlaps(t *testing.T) {
	filter := NewSameRuleGroupFilter()
	require.True(t, filter.Overlaps(makeRuleMatch(10, 20), makeRuleMatch(10, 20)))
	require.True(t, filter.Overlaps(makeRuleMatch(10, 20), makeRuleMatch(5, 11)))
	require.True(t, filter.Overlaps(makeRuleMatch(10, 20), makeRuleMatch(19, 21)))
	require.True(t, filter.Overlaps(makeRuleMatch(10, 20), makeRuleMatch(11, 19)))
	require.True(t, filter.Overlaps(makeRuleMatch(10, 20), makeRuleMatch(1, 10)))
	require.True(t, filter.Overlaps(makeRuleMatch(10, 20), makeRuleMatch(19, 20)))

	require.False(t, filter.Overlaps(makeRuleMatch(10, 20), makeRuleMatch(21, 30)))
	require.False(t, filter.Overlaps(makeRuleMatch(10, 20), makeRuleMatch(1, 9)))
}

func makeRuleMatch(fromPos, toPos int) *RuleMatch {
	return NewRuleMatch(NewFakeRule(""), nil, fromPos, toPos, "FakeMatch1")
}
