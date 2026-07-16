package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.CleanOverlappingFilterTest.

func myPriority(id string) int {
	switch id {
	case "P3_RULE":
		return 3
	case "P2_RULE", "P2_PREMIUM_RULE", "COMMA_HIGH_PRIORITY", "MISSING_THE_HIGH_PRIORITY":
		return 2
	case "P1_RULE", "P1_RULE_B", "P1_PREMIUM_RULE", "COMMA_LOW_PRIORITY", "COMMA_LOW_PRIORITY2", "MISSING_THE_LOW_PRIORITY":
		return 1
	case "MISC":
		return 0
	case "FAKE-RULE":
		return 0
	default:
		panic("No priority defined for " + id)
	}
}

func newCleanFilter() *CleanOverlappingFilter {
	return NewCleanOverlappingFilter(myPriority, true)
}

func emptySentence() *languagetool.AnalyzedSentence {
	return languagetool.AnalyzePlain("")
}

func TestCleanOverlappingFilter_Filter(t *testing.T) {
	filter := newCleanFilter()
	sentence := emptySentence()

	require.Len(t, filter.Filter(nil), 0)
	require.Len(t, filter.Filter([]*RuleMatch{}), 0)

	matches2 := []*RuleMatch{
		NewRuleMatch(NewFakeRule(""), sentence, 0, 10, "msg"),
		NewRuleMatch(NewFakeRule(""), sentence, 11, 20, "msg"),
	}
	require.Len(t, filter.Filter(matches2), 2)

	matches3 := []*RuleMatch{
		NewRuleMatch(NewFakeRule(""), sentence, 0, 10, "msg"),
		NewRuleMatch(NewFakeRule(""), sentence, 10, 20, "msg"),
	}
	require.Len(t, filter.Filter(matches3), 2)

	matches4 := filter.Filter([]*RuleMatch{
		NewRuleMatch(NewFakeRule("P1_RULE"), sentence, 0, 10, "msg1"),
		NewRuleMatch(NewFakeRule("P1_RULE_B"), sentence, 9, 20, "msg2"),
	})
	require.Len(t, matches4, 1)
	require.Equal(t, "P1_RULE_B", ruleIDOf(matches4[0].Rule))

	matches5 := filter.Filter([]*RuleMatch{
		NewRuleMatch(NewFakeRule("P1_RULE_B"), sentence, 0, 10, "msg1"),
		NewRuleMatch(NewFakeRule("P1_RULE"), sentence, 9, 20, "msg2"),
	})
	require.Len(t, matches5, 1)
	require.Equal(t, "P1_RULE", ruleIDOf(matches5[0].Rule))

	matches6 := filter.Filter([]*RuleMatch{
		NewRuleMatch(NewFakeRule("P1_RULE"), sentence, 0, 10, "msg1"),
		NewRuleMatch(NewFakeRule("P2_RULE"), sentence, 9, 20, "msg2"),
	})
	require.Len(t, matches6, 1)
	require.Equal(t, "P2_RULE", ruleIDOf(matches6[0].Rule))

	matches7 := filter.Filter([]*RuleMatch{
		NewRuleMatch(NewFakeRule("P2_RULE"), sentence, 0, 10, "msg1"),
		NewRuleMatch(NewFakeRule("P1_RULE"), sentence, 9, 20, "msg2"),
	})
	require.Len(t, matches7, 1)
	require.Equal(t, "P2_RULE", ruleIDOf(matches7[0].Rule))

	matches8 := filter.Filter([]*RuleMatch{
		NewRuleMatch(NewFakeRule("P2_PREMIUM_RULE"), sentence, 0, 10, "msg1"),
		NewRuleMatch(NewFakeRule("P1_RULE"), sentence, 9, 20, "msg2"),
	})
	require.Len(t, matches8, 1)
	require.Equal(t, "P1_RULE", ruleIDOf(matches8[0].Rule))

	matches8b := filter.Filter([]*RuleMatch{
		NewRuleMatch(NewFakeRule("P1_RULE"), sentence, 0, 10, "msg2"),
		NewRuleMatch(NewFakeRule("P2_PREMIUM_RULE"), sentence, 9, 20, "msg1"),
	})
	require.Len(t, matches8b, 1)
	require.Equal(t, "P1_RULE", ruleIDOf(matches8b[0].Rule))

	matches9 := filter.Filter([]*RuleMatch{
		NewRuleMatch(NewFakeRuleWithTag("P2_PREMIUM_RULE", TagPicky), sentence, 0, 10, "msg1"),
		NewRuleMatch(NewFakeRule("P1_RULE"), sentence, 9, 20, "msg2"),
	})
	require.Len(t, matches9, 1)
	require.Equal(t, "P1_RULE", ruleIDOf(matches9[0].Rule))

	matches10 := filter.Filter([]*RuleMatch{
		NewRuleMatch(NewFakeRule("P1_RULE"), sentence, 0, 10, "msg2"),
		NewRuleMatch(NewFakeRuleWithTag("P2_PREMIUM_RULE", TagPicky), sentence, 9, 20, "msg1"),
	})
	require.Len(t, matches10, 1)
	require.Equal(t, "P1_RULE", ruleIDOf(matches10[0].Rule))

	matches11 := filter.Filter([]*RuleMatch{
		NewRuleMatch(NewFakeRuleWithTag("P2_RULE", TagPicky), sentence, 0, 10, "msg1"),
		NewRuleMatch(NewFakeRule("P1_RULE"), sentence, 9, 20, "msg2"),
	})
	require.Len(t, matches11, 1)
	require.Equal(t, "P1_RULE", ruleIDOf(matches11[0].Rule))

	matches12 := filter.Filter([]*RuleMatch{
		NewRuleMatch(NewFakeRule("P1_RULE"), sentence, 0, 10, "msg2"),
		NewRuleMatch(NewFakeRuleWithTag("P2_RULE", TagPicky), sentence, 9, 20, "msg1"),
	})
	require.Len(t, matches12, 1)
	require.Equal(t, "P1_RULE", ruleIDOf(matches12[0].Rule))

	// juxtaposed matches, comma in the same place
	rm1 := NewRuleMatch(NewFakeRule("COMMA_LOW_PRIORITY"), sentence, 5, 10, "msg1")
	rm1.SetSuggestedReplacement("right,")
	rm2 := NewRuleMatch(NewFakeRule("COMMA_HIGH_PRIORITY"), sentence, 10, 15, "msg2")
	rm2.SetSuggestedReplacement(", left")
	matches14 := filter.Filter([]*RuleMatch{rm1, rm2})
	require.Len(t, matches14, 1)
	require.Equal(t, "COMMA_HIGH_PRIORITY", ruleIDOf(matches14[0].Rule))

	rm3 := NewRuleMatch(NewFakeRule("COMMA_HIGH_PRIORITY"), sentence, 5, 10, "msg1")
	rm3.SetSuggestedReplacement("right,")
	rm4 := NewRuleMatch(NewFakeRule("COMMA_LOW_PRIORITY"), sentence, 10, 15, "msg2")
	rm4.SetSuggestedReplacement(", left")
	matches15 := filter.Filter([]*RuleMatch{rm3, rm4})
	require.Len(t, matches15, 1)
	require.Equal(t, "COMMA_HIGH_PRIORITY", ruleIDOf(matches15[0].Rule))

	rm5 := NewRuleMatch(NewFakeRule("COMMA_LOW_PRIORITY2"), sentence, 5, 10, "msg1")
	rm5.SetSuggestedReplacement("right,")
	rm6 := NewRuleMatch(NewFakeRule("COMMA_LOW_PRIORITY"), sentence, 10, 15, "msg2")
	rm6.SetSuggestedReplacement(", left")
	matches16 := filter.Filter([]*RuleMatch{rm5, rm6})
	require.Len(t, matches16, 1)
	require.Equal(t, "COMMA_LOW_PRIORITY", ruleIDOf(matches16[0].Rule))

	// same suggestion for the same place
	rm7 := NewRuleMatch(NewFakeRule("MISSING_THE_HIGH_PRIORITY"), sentence, 5, 10, "msg1")
	rm7.SetSuggestedReplacement("of the")
	rm8 := NewRuleMatch(NewFakeRule("MISSING_THE_LOW_PRIORITY"), sentence, 11, 15, "msg2")
	rm8.SetSuggestedReplacement("the provisions")
	matches17 := filter.Filter([]*RuleMatch{rm7, rm8})
	require.Len(t, matches17, 1)
	require.Equal(t, "MISSING_THE_HIGH_PRIORITY", ruleIDOf(matches17[0].Rule))

	require.Panics(t, func() {
		filter.Filter([]*RuleMatch{
			NewRuleMatch(NewFakeRule("P1_RULE"), sentence, 11, 12, "msg2"),
			NewRuleMatch(NewFakeRuleWithTag("P2_RULE", TagPicky), sentence, 9, 10, "msg1"),
		})
	})
}
