package tools

// Twin of LtThreadPoolFactoryTest (non-@Ignore cases)
import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLtThreadPoolFactory_CachedThreadPoolTest(t *testing.T) {
	t.Cleanup(ResetLtThreadPoolFactory)
	a := CreateFixedThreadPoolExecutor("Test-Pool-cached", 10, 20, true)
	b := DefaultLtThreadPoolFactory.Get("Test-Pool-cached")
	require.NotNil(t, b)
	require.Same(t, a, b)
	// execute a task
	var n atomic.Int64
	require.True(t, a.Execute(func() { n.Add(1) }))
	require.Eventually(t, func() bool { return n.Load() == 1 }, time.Second, 5*time.Millisecond)
}

func TestLtThreadPoolFactory_NotcachedThreadPoolTest(t *testing.T) {
	t.Cleanup(ResetLtThreadPoolFactory)
	a := CreateFixedThreadPoolExecutor("Test-Pool-notCached", 10, 20, false)
	// non-reuse pools are not registered → Get is nil (Java returns defaultPool)
	require.Nil(t, DefaultLtThreadPoolFactory.Get("Test-Pool-notCached"))
	require.NotNil(t, a)
}
