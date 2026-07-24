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

// Twin of Java sentence.substring(from,to): multi-byte BMP must use UTF-16 indices.
// "café X" — café = 4 UTF-16 units; space at 4; X at 5. Replacing X (5..6) with "oops"
// yields "café oops" not a garbled UTF-8 slice.
func TestSuppressIfAnyRuleMatchesFilter_UTF16Substring(t *testing.T) {
	var saw string
	f := NewSuppressIfAnyRuleMatchesFilter(func(ruleID, sent string) []MatchSpan {
		saw = sent
		if ruleID == "BAD" && strings.Contains(sent, "oops") {
			// overlap original span 5..6
			return []MatchSpan{{From: 5, To: 9}}
		}
		return nil
	})
	// "café X" UTF-16: 0..3 café, 4 space, 5 X
	require.True(t, f.ShouldSuppress("café X", 5, 6, []string{"oops"}, "BAD"))
	require.Equal(t, "café oops", saw)

	// Non-BMP: "😀X" = hi, lo, X → X at UTF-16 index 2
	saw = ""
	f2 := NewSuppressIfAnyRuleMatchesFilter(func(ruleID, sent string) []MatchSpan {
		saw = sent
		return nil
	})
	require.False(t, f2.ShouldSuppress("😀X", 2, 3, []string{"Y"}, "OTHER"))
	require.Equal(t, "😀Y", saw)
}
