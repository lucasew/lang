package rules

import (
	"context"
	"testing"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRemoteRule_Match(t *testing.T) {
	cfg := NewRemoteRuleConfig()
	cfg.RuleID = "TEST_REMOTE"
	r := NewRemoteRule("en", cfg)
	r.Execute = func(sentences []*languagetool.AnalyzedSentence) *RemoteRuleResult {
		return &RemoteRuleResult{Matches: []*RuleMatch{
			NewRuleMatch(r, sentences[0], 0, 1, "remote hit"),
		}}
	}
	sent := languagetool.AnalyzePlain("hi")
	ms := r.MatchRemote([]*languagetool.AnalyzedSentence{sent})
	require.Len(t, ms, 1)
	require.Equal(t, "remote hit", ms[0].Message)
}

func TestRemoteRule_Timeout(t *testing.T) {
	// Config-driven timeout duration
	d := RemoteTimeoutMilliseconds(50, 1.0, 10)
	require.Equal(t, 60*time.Millisecond, d)
	_, err := RunWithTimeout(context.Background(), 20*time.Millisecond, func(ctx context.Context) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(200 * time.Millisecond):
			return 1, nil
		}
	})
	require.Error(t, err)
}

func TestRemoteRule_FailedRequests(t *testing.T) {
	cfg := NewRemoteRuleConfig()
	cfg.RuleID = "FAIL"
	cfg.FailureRateThreshold = 50
	r := NewRemoteRule("en", cfg)
	r.Execute = nil // no backend → empty
	require.Empty(t, r.MatchRemote([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("x")}))
	// TestRemoteRule Fallback path
	tr := NewTestRemoteRule(cfg)
	fb := tr.Fallback([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("x")})
	require.NotNil(t, fb)
}

func TestRemoteRule_AbFlags(t *testing.T) {
	cfg := NewRemoteRuleConfig()
	cfg.Options = map[string]string{"abTest": "groupA"}
	require.Equal(t, "groupA", cfg.Options["abTest"])
}

func TestRemoteRule_ThirdPartyAI(t *testing.T) {
	cfg := NewRemoteRuleConfig()
	cfg.RuleID = "AI_PRIMARY"
	cfg.Options = map[string]string{RemoteOptionThirdPartyAI: "true"}
	require.Equal(t, "true", cfg.Options[RemoteOptionThirdPartyAI])
}

func TestRemoteRule_ThirdPartyAIFallback(t *testing.T) {
	RemoteRuleFallbackInstance.Clear()
	primary := NewRemoteRuleConfig()
	primary.RuleID = "AI_RULE"
	primary.Options = map[string]string{RemoteOptionFallbackRule: "LOCAL_RULE", RemoteOptionThirdPartyAI: "true"}
	fallback := NewRemoteRuleConfig()
	fallback.RuleID = "LOCAL_RULE"
	RemoteRuleFallbackInstance.InitForTests([]*RemoteRuleConfig{primary, fallback})
	require.True(t, RemoteRuleFallbackInstance.HasFallback("AI_RULE"))
	require.Equal(t, "LOCAL_RULE", RemoteRuleFallbackInstance.GetFallback("AI_RULE").GetRuleID())
}
