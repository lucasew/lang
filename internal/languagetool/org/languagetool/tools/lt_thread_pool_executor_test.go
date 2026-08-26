package tools

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLtThreadPoolExecutor(t *testing.T) {
	p := NewLtThreadPoolExecutor("test", 2, 8)
	p.Start()
	defer p.Shutdown()

	var n atomic.Int64
	for i := 0; i < 5; i++ {
		ok := p.Execute(func() {
			n.Add(1)
		})
		require.True(t, ok)
	}
	require.Eventually(t, func() bool { return n.Load() == 5 }, time.Second, 10*time.Millisecond)
	require.Equal(t, "test", p.Name())
	require.Equal(t, 8, p.MaxQueueSize())
}
