package remote

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSentenceAnnotatorCacheKey(t *testing.T) {
	k1 := CacheKey("Hello", "en")
	k2 := CacheKey("Hello", "en")
	k3 := CacheKey("Hello", "de")
	require.Equal(t, k1, k2)
	require.NotEqual(t, k1, k3)
	require.Equal(t, "2006-01-02", TimestampPrefix(time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC)))

	cfg := DefaultAnnotatorConfig()
	a := NewSentenceAnnotator(cfg)
	require.NotNil(t, a.Client)
	// cache hit without network: seed cache
	a.Cache[CacheKey("x", "en-US")] = nil
	m, err := a.AnnotateSentence("x")
	require.NoError(t, err)
	require.Nil(t, m)
}
