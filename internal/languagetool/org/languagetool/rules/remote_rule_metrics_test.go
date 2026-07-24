package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoteRuleMetrics(t *testing.T) {
	m := NewRemoteRuleMetrics()
	m.RecordRequest("R1", 0.1, 100, RemoteResultSuccess)
	m.RecordRequest("R1", 0.2, 50, RemoteResultTimeout)
	require.Equal(t, 1, m.RequestCounts["R1"][RemoteResultSuccess])
	require.Equal(t, 1, m.RequestCounts["R1"][RemoteResultTimeout])
	m.RecordWait("en", 500)
	require.InDelta(t, 0.5, m.WaitSeconds["en"][0], 1e-9)
}
