package bert

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoteLanguageModel(t *testing.T) {
	m := NewRemoteLanguageModel(EditDistanceScorer)
	req := NewRequest("I has a cat", 2, 5, []string{"have", "had", "xx"})
	scores, err := m.Score(req)
	require.NoError(t, err)
	require.Len(t, scores, 3)
	// "has" → "had"/"have" closer than "xx"
	require.Greater(t, scores[0], scores[2])

	// cache
	scores2, err := m.Score(req)
	require.NoError(t, err)
	require.Equal(t, scores, scores2)

	batch, err := m.BatchScore([]Request{req, NewRequest("a", 0, 1, []string{"a", "b"})})
	require.NoError(t, err)
	require.Len(t, batch, 2)
	m.Shutdown()
}
