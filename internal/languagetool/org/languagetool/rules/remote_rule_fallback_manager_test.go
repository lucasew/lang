package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoteRuleFallbackManager(t *testing.T) {
	RemoteRuleFallbackInstance.Clear()
	primary := NewRemoteRuleConfig()
	primary.RuleID = "AI_RULE"
	primary.Options["fallbackRuleId"] = "LOCAL_RULE"
	fallback := NewRemoteRuleConfig()
	fallback.RuleID = "LOCAL_RULE"
	RemoteRuleFallbackInstance.InitForTests([]*RemoteRuleConfig{primary, fallback})
	require.True(t, RemoteRuleFallbackInstance.HasFallback("AI_RULE"))
	require.Equal(t, "LOCAL_RULE", RemoteRuleFallbackInstance.GetFallback("AI_RULE").GetRuleID())
	require.False(t, RemoteRuleFallbackInstance.HasFallback("OTHER"))
}
