package tools

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker(t *testing.T) {
	t.Cleanup(ResetCircuitBreakers)
	b := NewCircuitBreaker("remote", 2, 20*time.Millisecond)
	require.True(t, b.Allow())
	b.OnFailure()
	require.True(t, b.Allow())
	b.OnFailure()
	require.False(t, b.Allow())
	require.Equal(t, CircuitOpen, b.State())
	time.Sleep(25 * time.Millisecond)
	require.Equal(t, CircuitHalfOpen, b.State())
	require.True(t, b.Allow())
	b.OnSuccess()
	require.Equal(t, CircuitClosed, b.State())

	reg := CircuitBreakerRegistry()
	a := reg.GetOrCreate("x")
	a2 := reg.GetOrCreate("x")
	require.Same(t, a, a2)
}
