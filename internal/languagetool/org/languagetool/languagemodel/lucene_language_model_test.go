package languagemodel

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLuceneLanguageModelMergeCounts(t *testing.T) {
	a := NewIndexLanguageModel("a", &MapCountProvider{
		Counts: map[string]int64{"hello": 3, "hello\x00world": 1},
		Total:  10,
	})
	b := NewIndexLanguageModel("b", &MapCountProvider{
		Counts: map[string]int64{"hello": 2},
		Total:  5,
	})
	lm := NewLuceneLanguageModelFromIndexes([]*IndexLanguageModel{a, b})
	require.Equal(t, int64(5), lm.GetCountToken("hello"))
	require.Equal(t, int64(15), lm.GetTotalTokenCount())
	p := lm.GetPseudoProbability([]string{"hello"})
	require.Greater(t, p.GetProb(), 0.0)
	require.NoError(t, lm.Close())
}

func TestLuceneSearcherCache(t *testing.T) {
	ClearLuceneSearcherCache()
	c := &MapCountProvider{Counts: map[string]int64{"a": 3}, Total: 3}
	s1 := GetCachedLuceneSearcher("/tmp/idx", c)
	s2 := GetCachedLuceneSearcher("/tmp/idx", nil)
	require.Same(t, s1, s2)
	require.Equal(t, int64(3), s1.GetCountToken("a"))
	lm := NewLuceneSingleIndexLanguageModelMap("/tmp/idx", c)
	require.Equal(t, int64(3), lm.GetCountToken("a"))
}
