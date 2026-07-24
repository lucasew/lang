package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGRPCRule(t *testing.T) {
	cfg := NewRemoteRuleConfig()
	cfg.RuleID = "EXAMPLE"
	cfg.Type = GRPCRuleConfigType
	g := CreateGRPCRule("en", cfg, "EXAMPLE", "desc", map[string]string{"EXAMPLE": "fix me"})
	g.MatchSentences = func(sentences []*languagetool.AnalyzedSentence) [][]*RuleMatch {
		out := make([][]*RuleMatch, len(sentences))
		for i, s := range sentences {
			out[i] = []*RuleMatch{NewRuleMatch(NewFakeRule("EXAMPLE"), s, 0, 1, "")}
		}
		return out
	}
	sent := languagetool.AnalyzePlain("x")
	ms := g.MatchRemote([]*languagetool.AnalyzedSentence{sent})
	require.Len(t, ms, 1)
}

func TestGRPCPostProcessing(t *testing.T) {
	t.Cleanup(ResetGRPCPostProcessing)
	cfg := NewRemoteRuleConfig()
	cfg.RuleID = "POST1"
	cfg.Type = GRPCPostConfigType
	ConfigureGRPCPostProcessing("en", []*RemoteRuleConfig{cfg})
	list := GetGRPCPostProcessing("en")
	require.Len(t, list, 1)
	list[0].Process = func(_ []*languagetool.AnalyzedSentence, matches []*RuleMatch) []*RuleMatch {
		return nil // drop all
	}
	sent := languagetool.AnalyzePlain("hi")
	m := NewRuleMatch(NewFakeRule("X"), sent, 0, 1, "m")
	out := list[0].Apply([]*languagetool.AnalyzedSentence{sent}, []*RuleMatch{m})
	require.Empty(t, out)
}
