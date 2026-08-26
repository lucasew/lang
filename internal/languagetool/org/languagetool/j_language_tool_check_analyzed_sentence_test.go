package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin surface of checkAnalyzedSentence: registered sentence checkers only;
// ONLYPARA / text-level-only modes return empty.
func TestCheckAnalyzedSentence_RunsCheckers(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("DEMO_DUP", func(s *AnalyzedSentence) []LocalMatch {
		// flag any "xx xx" style by crude surface
		if s == nil {
			return nil
		}
		txt := s.GetText()
		if len(txt) >= 2 {
			return []LocalMatch{{FromPos: 0, ToPos: 1, RuleID: "DEMO_DUP", Message: "demo"}}
		}
		return nil
	})
	s := AnalyzePlain("hello")
	ms := lt.CheckAnalyzedSentence(s)
	require.NotEmpty(t, ms)
	require.Equal(t, "DEMO_DUP", ms[0].RuleID)

	// ONLYPARA skips sentence rules (Java checkAnalyzedSentence)
	lt.ParaMode = ParagraphOnlyPara
	require.Empty(t, lt.CheckAnalyzedSentence(s))
}

func TestCheckAnalyzedSentence_DisabledFiltered(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("OFF_ME", func(s *AnalyzedSentence) []LocalMatch {
		return []LocalMatch{{FromPos: 0, ToPos: 1, RuleID: "OFF_ME", Message: "x"}}
	})
	lt.DisableRule("OFF_ME")
	s := AnalyzePlain("ab")
	require.Empty(t, lt.CheckAnalyzedSentence(s))
}
