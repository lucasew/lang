package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRemoteRuleConfigOptions(t *testing.T) {
	cfg := NewRemoteRuleConfig()
	cfg.RuleID = "REMOTE_X"
	cfg.Options = map[string]string{
		"filterMatches": "true",
		"premium":       "true",
	}
	r := NewRemoteRule("en", cfg)
	require.Equal(t, "REMOTE_X", r.GetID())
	require.True(t, r.FilterMatches)
	require.True(t, r.Premium)
	// execute
	r.Execute = func(sentences []*languagetool.AnalyzedSentence) *RemoteRuleResult {
		return &RemoteRuleResult{Matches: []*RuleMatch{NewRuleMatch(r, sentences[0], 0, 1, "remote")}}
	}
	sent := languagetool.AnalyzePlain("hi")
	ms := r.MatchRemote([]*languagetool.AnalyzedSentence{sent})
	require.Len(t, ms, 1)
}
