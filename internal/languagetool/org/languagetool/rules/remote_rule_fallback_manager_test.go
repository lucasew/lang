package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoteRuleFallbackManager(t *testing.T) {
	RemoteRuleFallbackInstance.Clear()
	primary := NewRemoteRuleConfig()
	primary.RuleID = "AI_RULE"
	primary.Options = map[string]string{RemoteOptionFallbackRule: "LOCAL_RULE"}
	fallback := NewRemoteRuleConfig()
	fallback.RuleID = "LOCAL_RULE"
	RemoteRuleFallbackInstance.InitForTests([]*RemoteRuleConfig{primary, fallback})
	require.True(t, RemoteRuleFallbackInstance.HasFallback("AI_RULE"))
	require.Equal(t, "LOCAL_RULE", RemoteRuleFallbackInstance.GetFallback("AI_RULE").GetRuleID())
	require.False(t, RemoteRuleFallbackInstance.HasFallback("OTHER"))
	// GetInhouseFallback rejects third-party AI fallbacks
	require.Equal(t, "LOCAL_RULE", RemoteRuleFallbackInstance.GetInhouseFallback("AI_RULE").GetRuleID())
	fallback.Options = map[string]string{RemoteOptionThirdPartyAI: "true"}
	require.Nil(t, RemoteRuleFallbackInstance.GetInhouseFallback("AI_RULE"))
}

type fakeRemoteRule struct {
	id, fb  string
	state   CircuitBreakerState
}

func (f fakeRemoteRule) GetID() string                       { return f.id }
func (f fakeRemoteRule) CircuitBreakerState() CircuitBreakerState { return f.state }
func (f fakeRemoteRule) GetFallbackRuleId() string           { return f.fb }

func TestIsRuleOrFallbackAvailable(t *testing.T) {
	m := &RemoteRuleFallbackManager{}
	a := fakeRemoteRule{id: "A", state: CircuitClosed}
	require.Equal(t, "A", m.IsRuleOrFallbackAvailable(a, map[string]RemoteRuleAvailability{"A": a}))
	// open with fallback closed
	b := fakeRemoteRule{id: "B", fb: "C", state: CircuitOpen}
	c := fakeRemoteRule{id: "C", state: CircuitClosed}
	require.Equal(t, "C", m.IsRuleOrFallbackAvailable(b, map[string]RemoteRuleAvailability{"B": b, "C": c}))
	// circular
	x := fakeRemoteRule{id: "X", fb: "Y", state: CircuitOpen}
	y := fakeRemoteRule{id: "Y", fb: "X", state: CircuitOpen}
	require.Equal(t, "", m.IsRuleOrFallbackAvailable(x, map[string]RemoteRuleAvailability{"X": x, "Y": y}))
}
