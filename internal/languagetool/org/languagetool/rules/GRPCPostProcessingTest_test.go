package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGRPCPostProcessing_RuleMatchModification(t *testing.T) {
	t.Cleanup(ResetGRPCPostProcessing)
	cfg := NewRemoteRuleConfig()
	cfg.RuleID = "POST_MOD"
	cfg.Type = GRPCPostConfigType
	ConfigureGRPCPostProcessing("en", []*RemoteRuleConfig{cfg})
	list := GetGRPCPostProcessing("en")
	require.Len(t, list, 1)
	list[0].Process = func(_ []*languagetool.AnalyzedSentence, matches []*RuleMatch) []*RuleMatch {
		for _, m := range matches {
			if m != nil {
				m.Message = "modified"
			}
		}
		return matches
	}
	sent := languagetool.AnalyzePlain("hi")
	m := NewRuleMatch(NewFakeRule("X"), sent, 0, 1, "orig")
	out := list[0].Apply([]*languagetool.AnalyzedSentence{sent}, []*RuleMatch{m})
	require.Len(t, out, 1)
	require.Equal(t, "modified", out[0].Message)
}

func TestGRPCPostProcessing_TagEnums(t *testing.T) {
	// Tag constants used by picky/post rules
	require.Equal(t, Tag("picky"), TagPicky)
}

func TestGRPCPostProcessing_MatchTypeEnums(t *testing.T) {
	// ITS issue types surface for match classification
	require.Equal(t, ITSIssueType("misspelling"), ITSMisspelling)
	require.Equal(t, ITSIssueType("grammar"), ITSGrammar)
	require.Equal(t, ITSIssueType("style"), ITSStyle)
}

func TestGRPCPostProcessing_SuggestionTypeEnums(t *testing.T) {
	require.Equal(t, SuggestionTypeDefault, SuggestionType("Default"))
	require.Equal(t, SuggestionTypeTranslation, SuggestionType("Translation"))
	require.Equal(t, SuggestionTypeCurated, SuggestionType("Curated"))
	s := NewSuggestedReplacement("ok")
	s.SetType(SuggestionTypeCurated)
	require.Equal(t, SuggestionTypeCurated, s.GetType())
}
