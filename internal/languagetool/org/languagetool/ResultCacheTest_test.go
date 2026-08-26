package languagetool

// Twin of ResultCacheTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResultCache_SimpleInputSentenceCache(t *testing.T) {
	c := NewResultCache(10)
	require.Equal(t, int64(0), c.HitCount())
	// empty cache: Java reports hitRate 1.0 before any request; we report 0 until first lookup
	key := NewSimpleInputSentence("foo", "de")
	_, ok := c.GetSentenceIfPresent(key)
	require.False(t, ok)
	sent := AnalyzePlain("foo")
	c.PutSentence(key, sent)
	got, ok := c.GetSentenceIfPresent(key)
	require.True(t, ok)
	require.Equal(t, sent.GetText(), got.GetText())
	// different language / text → miss
	_, ok = c.GetSentenceIfPresent(NewSimpleInputSentence("foo", "de-DE"))
	require.False(t, ok)
	_, ok = c.GetSentenceIfPresent(NewSimpleInputSentence("foo", "en"))
	require.False(t, ok)
	_, ok = c.GetSentenceIfPresent(NewSimpleInputSentence("foo bar", "de"))
	require.False(t, ok)
	require.GreaterOrEqual(t, c.HitCount(), int64(1))
	require.Greater(t, c.RequestCount(), c.HitCount())
	require.Greater(t, c.HitRate(), 0.0)
	require.LessOrEqual(t, c.HitRate(), 1.0)
}

func TestResultCache_InputSentenceCache(t *testing.T) {
	c := NewResultCache(100)
	analyzed := AnalyzePlain("foo")
	key := NewInputSentence(analyzed, "de", "", nil, nil, nil, nil, nil, nil, "ALL", LevelDefault, nil, nil)
	keySame := NewInputSentence(AnalyzePlain("foo"), "de", "", nil, nil, nil, nil, nil, nil, "ALL", LevelDefault, nil, nil)
	_, ok := c.GetMatchesIfPresent(key)
	require.False(t, ok)
	c.PutMatches(key, []string{"m1"})
	got, ok := c.GetMatchesIfPresent(key)
	require.True(t, ok)
	require.Equal(t, []string{"m1"}, got)
	// same text+lang+mode → hit
	got2, ok := c.GetMatchesIfPresent(keySame)
	require.True(t, ok)
	require.Equal(t, got, got2)
	// different text / lang / mother tongue / disabled rules → miss
	_, ok = c.GetMatchesIfPresent(NewInputSentence(AnalyzePlain("foo bar"), "de", "", nil, nil, nil, nil, nil, nil, "ALL", LevelDefault, nil, nil))
	require.False(t, ok)
	_, ok = c.GetMatchesIfPresent(NewInputSentence(AnalyzePlain("foo"), "en", "", nil, nil, nil, nil, nil, nil, "ALL", LevelDefault, nil, nil))
	require.False(t, ok)
	_, ok = c.GetMatchesIfPresent(NewInputSentence(AnalyzePlain("foo"), "de", "en", nil, nil, nil, nil, nil, nil, "ALL", LevelDefault, nil, nil))
	require.False(t, ok)
}

func TestResultCache_RemoteMatches(t *testing.T) {
	c := NewResultCache(10)
	_, ok := c.GetRemoteMatchesIfPresent("Foo. ", "TEST_REMOTE_RULE")
	require.False(t, ok)
	c.PutRemoteMatches("Foo. ", "TEST_REMOTE_RULE", []int{0, 1})
	got, ok := c.GetRemoteMatchesIfPresent("Foo. ", "TEST_REMOTE_RULE")
	require.True(t, ok)
	require.Equal(t, []int{0, 1}, got)
	// different rule id or text → miss
	_, ok = c.GetRemoteMatchesIfPresent("Foo. ", "OTHER")
	require.False(t, ok)
	_, ok = c.GetRemoteMatchesIfPresent("Bar.", "TEST_REMOTE_RULE")
	require.False(t, ok)
}

func TestResultCache_ZeroMaxSize(t *testing.T) {
	c := NewResultCache(0)
	c.PutSentence(NewSimpleInputSentence("x", "en"), AnalyzePlain("x"))
	_, ok := c.GetSentenceIfPresent(NewSimpleInputSentence("x", "en"))
	require.False(t, ok)
}
