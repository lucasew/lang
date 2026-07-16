package server

// Twin of RequestLimiterTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestLimiter_IsAccessOkay(t *testing.T) {
	l := NewRequestLimiter(2, 60)
	require.True(t, l.Allow("1.1.1.1"))
	require.True(t, l.Allow("1.1.1.1"))
	require.False(t, l.Allow("1.1.1.1"))
	// different IP independent
	require.True(t, l.Allow("2.2.2.2"))
	require.Equal(t, 2, l.Count("1.1.1.1"))
}

func TestRequestLimiter_IsAccessOkayWithFingerprintDisabled(t *testing.T) {
	// Fingerprint is not modeled; IP-only limiter still works.
	l := NewRequestLimiter(1, 60)
	require.True(t, l.Allow("10.0.0.1"))
	require.False(t, l.Allow("10.0.0.1"))
}

func TestRequestLimiter_NilAllows(t *testing.T) {
	var l *RequestLimiter
	require.True(t, l.Allow("any"))
}
