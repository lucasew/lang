package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoteRuleConfig_Parse(t *testing.T) {
	js := `[
  {"ruleId":"AI_HINT","url":"https://example.com","type":"gpt","language":"en.*","options":{"premium":"true"}},
  {"ruleId":"OTHER","url":"http://x","port":8080,"type":"gpt"}
]`
	list, err := ParseRemoteRuleConfigs(strings.NewReader(js))
	require.NoError(t, err)
	require.Len(t, list, 2)
	require.Equal(t, "AI_HINT", list[0].GetRuleID())
	require.Equal(t, RemoteDefaultPort, list[0].GetPort())
	require.Equal(t, 8080, list[1].GetPort())
	require.Equal(t, list[0], GetRelevantRemoteRuleConfig("AI_HINT", list))
	require.Nil(t, GetRelevantRemoteRuleConfig("NOPE", list))

	pred := IsRelevantRemoteRuleConfig("gpt", "en-US")
	require.True(t, pred(list[0]))
	require.True(t, pred(list[1])) // language empty → all
	predES := IsRelevantRemoteRuleConfig("gpt", "es")
	require.False(t, predES(list[0]))
	require.True(t, predES(list[1]))

	cp := CopyRemoteRuleConfig(list[0])
	cp.RuleID = "CHANGED"
	require.Equal(t, "AI_HINT", list[0].RuleID)
}
