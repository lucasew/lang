package rules

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Port of RemoteRuleTimeoutTest.testCancelThreads (unit surface; Java @Ignore interactive)
func TestRemoteRuleTimeout_CancelThreads(t *testing.T) {
	// Completes within timeout
	v, err := RunWithTimeout(context.Background(), 200*time.Millisecond, func(ctx context.Context) (int, error) {
		return 42, nil
	})
	require.NoError(t, err)
	require.Equal(t, 42, v)

	// Exceeds timeout
	_, err = RunWithTimeout(context.Background(), 30*time.Millisecond, func(ctx context.Context) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(500 * time.Millisecond):
			return 1, nil
		}
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "timeout")

	// Config-style duration
	d := RemoteTimeoutMilliseconds(50, 0.1, 100)
	require.Equal(t, 60*time.Millisecond, d)
}
