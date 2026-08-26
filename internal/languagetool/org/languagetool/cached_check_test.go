package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCachedCheck(t *testing.T) {
	cache := NewResultCache(100)
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
	src := "This is an test."
	m1 := CachedCheck(cache, lt, src)
	require.NotEmpty(t, m1)
	// second call hits cache
	m2 := CachedCheck(cache, lt, src)
	require.Equal(t, m1, m2)
	require.Greater(t, cache.HitCount(), int64(0))
	// disable changes key → miss, empty
	lt.DisableRule("EN_A_VS_AN")
	m3 := CachedCheck(cache, lt, src)
	require.Empty(t, m3)
}
