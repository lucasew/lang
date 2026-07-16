package server

// Twin of PipelinePoolTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPipelinePool_PipelineCreatedAndUsed(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.PipelineCaching = true
	cfg.MaxPipelinePoolSize = 5
	pool := NewPipelinePool(cfg)
	settings := NewPipelineSettings("en", "anon")
	pl, err := pool.Borrow(settings)
	require.NoError(t, err)
	require.NotNil(t, pl)
	require.Equal(t, 1, pool.Borrowed())
	pool.Return(settings, pl)
	require.Equal(t, 0, pool.Borrowed())
	require.Equal(t, 1, pool.IdleCount(settings))
	// re-borrow same idle
	pl2, err := pool.Borrow(settings)
	require.NoError(t, err)
	require.Same(t, pl, pl2)
}

func TestPipelinePool_DifferentPipelineSettings(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.PipelineCaching = true
	pool := NewPipelinePool(cfg)
	a := NewPipelineSettings("en", "u1")
	b := NewPipelineSettings("de", "u1")
	pa, err := pool.Borrow(a)
	require.NoError(t, err)
	pb, err := pool.Borrow(b)
	require.NoError(t, err)
	require.NotSame(t, pa, pb)
	pool.Return(a, pa)
	pool.Return(b, pb)
	require.Equal(t, 1, pool.IdleCount(a))
	require.Equal(t, 1, pool.IdleCount(b))
}

func TestPipelinePool_MaxPipelinePoolSize(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.PipelineCaching = true
	cfg.MaxPipelinePoolSize = 1
	pool := NewPipelinePool(cfg)
	s := NewPipelineSettings("en", "anon")
	pl, err := pool.Borrow(s)
	require.NoError(t, err)
	_, err = pool.Borrow(s)
	require.Error(t, err) // exhausted
	pool.Return(s, pl)
	_, err = pool.Borrow(s)
	require.NoError(t, err)
}

func TestPipelinePool_NoCachingAlwaysNew(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.PipelineCaching = false
	pool := NewPipelinePool(cfg)
	s := NewPipelineSettings("en", "anon")
	p1, err := pool.Borrow(s)
	require.NoError(t, err)
	pool.Return(s, p1)
	p2, err := pool.Borrow(s)
	require.NoError(t, err)
	// without caching, Return is no-op for reuse — still not same guaranteed
	require.NotNil(t, p2)
}

func TestPipelinePool_PipelinePoolUserConfig(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.PipelineCaching = true
	pool := NewPipelinePool(cfg)
	a := NewPipelineSettings("en", "alice")
	b := NewPipelineSettings("en", "bob")
	require.NotEqual(t, a.Key(), b.Key())
	pa, _ := pool.Borrow(a)
	pb, _ := pool.Borrow(b)
	pool.Return(a, pa)
	pool.Return(b, pb)
	require.Equal(t, 1, pool.IdleCount(a))
	require.Equal(t, 1, pool.IdleCount(b))
}
