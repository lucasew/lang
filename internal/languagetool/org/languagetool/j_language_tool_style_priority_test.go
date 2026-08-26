package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyRulePriorities_DefaultStyle(t *testing.T) {
	// Java English/Catalan getDefaultRulePriorityForStyle = -50 for ITS Style
	lt := NewJLanguageTool("en")
	lt.PriorityForId = func(id string) int { return 0 } // no map hits
	lt.DefaultRulePriorityForStyle = -50
	ms := []LocalMatch{
		{RuleID: "SOME_STYLEISH", IssueType: "style", Priority: 0},
		{RuleID: "SOME_GRAMMAR", IssueType: "grammar", Priority: 0},
		{RuleID: "MAPPED", IssueType: "style", Priority: 0},
	}
	// With PriorityForId returning 0 for all, style gets -50
	out := lt.applyRulePriorities(ms)
	require.Equal(t, -50, out[0].Priority)
	require.Equal(t, 0, out[1].Priority)

	// Rule id priority wins over style default
	lt.PriorityForId = func(id string) int {
		if id == "MAPPED" {
			return 10
		}
		return 0
	}
	ms2 := []LocalMatch{
		{RuleID: "MAPPED", IssueType: "style", Priority: 0},
		{RuleID: "OTHER_STYLE", IssueType: "Style", Priority: 0}, // case-insensitive
	}
	out2 := lt.applyRulePriorities(ms2)
	require.Equal(t, 10, out2[0].Priority)
	require.Equal(t, -50, out2[1].Priority)

	// Explicit Priority inject wins
	ms3 := []LocalMatch{{RuleID: "X", IssueType: "style", Priority: 7}}
	out3 := lt.applyRulePriorities(ms3)
	require.Equal(t, 7, out3[0].Priority)
}

func TestApplyRulePriorities_CategoryBeforeStyle(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.DefaultRulePriorityForStyle = -50
	lt.PriorityForId = func(id string) int {
		if id == "STYLE" {
			return -15 // category priority (e.g. German STYLE category)
		}
		return 0
	}
	ms := []LocalMatch{{RuleID: "FOO", CategoryID: "STYLE", IssueType: "style", Priority: 0}}
	out := lt.applyRulePriorities(ms)
	// category priority applies before style default
	require.Equal(t, -15, out[0].Priority)
}
