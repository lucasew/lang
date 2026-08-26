package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// CleanOverlapping runs inside Check after priorities (Java CleanOverlappingFilter).
func TestCheck_CleanOverlappingByPriority(t *testing.T) {
	lt := NewJLanguageTool("de")
	lt.AddRuleChecker("low", func(s *AnalyzedSentence) []LocalMatch {
		return []LocalMatch{{FromPos: 0, ToPos: 5, RuleID: "low", Priority: 1, Message: "low"}}
	})
	lt.AddRuleChecker("high", func(s *AnalyzedSentence) []LocalMatch {
		return []LocalMatch{{FromPos: 2, ToPos: 8, RuleID: "high", Priority: 10, Message: "high"}}
	})
	ms := lt.Check("abcdefgh")
	require.Len(t, ms, 1)
	require.Equal(t, "high", ms[0].RuleID)
}

func TestCheck_DisableCleanOverlapping(t *testing.T) {
	lt := NewJLanguageTool("de")
	lt.DisableCleanOverlapping()
	lt.AddRuleChecker("low", func(s *AnalyzedSentence) []LocalMatch {
		return []LocalMatch{{FromPos: 0, ToPos: 5, RuleID: "low", Priority: 1}}
	})
	lt.AddRuleChecker("high", func(s *AnalyzedSentence) []LocalMatch {
		return []LocalMatch{{FromPos: 2, ToPos: 8, RuleID: "high", Priority: 10}}
	})
	ms := lt.Check("abcdefgh")
	require.Len(t, ms, 2)
}

func TestCleanSameRuleGroupLocalMatches(t *testing.T) {
	in := []LocalMatch{
		{FromPos: 0, ToPos: 5, RuleID: "R1"},
		{FromPos: 3, ToPos: 8, RuleID: "R1"}, // same id overlap → drop second
		{FromPos: 10, ToPos: 12, RuleID: "R2"},
	}
	out := CleanSameRuleGroupLocalMatches(in)
	require.Len(t, out, 2)
	require.Equal(t, "R1", out[0].RuleID)
	require.Equal(t, 0, out[0].FromPos)
	require.Equal(t, "R2", out[1].RuleID)
}

func TestCheck_SameRuleGroupFilter(t *testing.T) {
	lt := NewJLanguageTool("de")
	lt.AddRuleChecker("R1", func(s *AnalyzedSentence) []LocalMatch {
		return []LocalMatch{
			{FromPos: 0, ToPos: 5, RuleID: "R1"},
			{FromPos: 3, ToPos: 8, RuleID: "R1"},
		}
	})
	ms := lt.Check("abcdefgh")
	require.Len(t, ms, 1)
	require.Equal(t, 0, ms[0].FromPos)
}
