package remote

// Twin of RemoteRuleMatchTest (http-client)
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoteRuleMatch_ToStringOutput(t *testing.T) {
	m := NewRemoteRuleMatch("ruleId", "ruleName", "msg", "context", 0, 0, 1)
	require.Equal(t, "ruleId@0-1", m.String())
}
