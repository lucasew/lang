package tools

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLtThreadPoolFactory(t *testing.T) {
	t.Cleanup(ResetLtThreadPoolFactory)
	a := CreateFixedThreadPoolExecutor("test-pool", 2, 8, true)
	b := CreateFixedThreadPoolExecutor("test-pool", 2, 8, true)
	require.Same(t, a, b)

	var n atomic.Int64
	require.True(t, a.Execute(func() { n.Add(1) }))
	require.Eventually(t, func() bool { return n.Load() == 1 }, time.Second, 5*time.Millisecond)

	c := CreateFixedThreadPoolExecutor("other", 1, 4, false)
	require.NotSame(t, a, c)
}
