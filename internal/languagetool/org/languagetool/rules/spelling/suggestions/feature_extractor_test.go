package suggestions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type mapLM map[string]float64

func (m mapLM) PseudoProbability(tokens []string) float64 {
	if len(tokens) == 0 {
		return 0
	}
	return m[tokens[0]]
}
func (m mapLM) Count(word string) int64 {
	if p, ok := m[word]; ok && p > 0 {
		return int64(p * 1000)
	}
	return 0
}

func TestFeatureExtractorOrder(t *testing.T) {
	e := NewSuggestionsOrdererFeatureExtractor(mapLM{"hello": 0.9, "hallo": 0.1})
	require.True(t, e.IsMlAvailable())
	got := OrderSuggestionsUsingModel(e, []string{"helo", "hello", "hallo"}, "hello", nil, 0)
	require.Equal(t, "hello", got[0])
	ordered, agg := e.ComputeFeatures([]string{"helo", "hello"}, "hello", nil, 0)
	require.Len(t, ordered, 2)
	require.Contains(t, agg, "top_levenshtein")
}

func TestJaroWinkler(t *testing.T) {
	require.InDelta(t, 1.0, jaroWinkler("abc", "abc"), 0.001)
	require.Greater(t, jaroWinkler("hello", "hallo"), jaroWinkler("hello", "xyz"))
}
