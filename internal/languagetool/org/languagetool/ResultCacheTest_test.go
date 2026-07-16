package languagetool

// Twin of ResultCacheTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResultCache_SimpleInputSentenceCache(t *testing.T) {
	c := NewResultCache(10)
	key := NewSimpleInputSentence("hello", "en")
	_, ok := c.GetSentenceIfPresent(key)
	require.False(t, ok)
	sent := AnalyzePlain("hello")
	c.PutSentence(key, sent)
	got, ok := c.GetSentenceIfPresent(key)
	require.True(t, ok)
	require.Equal(t, sent.GetText(), got.GetText())
	require.GreaterOrEqual(t, c.HitCount(), int64(1))
}

func TestResultCache_InputSentenceCache(t *testing.T) {
	c := NewResultCache(10)
	analyzed := AnalyzePlain("hi")
	key := NewInputSentence(analyzed, "en", "", nil, nil, nil, nil, nil, nil, "ALL", LevelDefault, nil, nil)
	_, ok := c.GetMatchesIfPresent(key)
	require.False(t, ok)
	c.PutMatches(key, []string{"m1"})
	got, ok := c.GetMatchesIfPresent(key)
	require.True(t, ok)
	require.Equal(t, []string{"m1"}, got)
}
