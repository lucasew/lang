package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestCounter(t *testing.T) {
	c := NewRequestCounter()
	require.Equal(t, 1, c.IncrementRequestCount())
	require.Equal(t, 2, c.IncrementRequestCount())
	require.Equal(t, 2, c.RequestCount())
	c.IncrementHandleCount("1.1.1.1", 10)
	c.IncrementHandleCount("1.1.1.1", 11)
	c.IncrementHandleCount("2.2.2.2", 12)
	require.Equal(t, 3, c.HandleCount())
	require.Equal(t, 2, c.DistinctIPs())
	c.DecrementHandleCount(11)
	require.Equal(t, 2, c.HandleCount())
	require.Equal(t, 2, c.DistinctIPs())
}
