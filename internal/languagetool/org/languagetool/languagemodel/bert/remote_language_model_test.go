package bert

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoteLanguageModel(t *testing.T) {
	var scoreCalls int32
	m := NewRemoteLanguageModel(func(req Request) ([]float64, error) {
		atomic.AddInt32(&scoreCalls, 1)
		return EditDistanceScorer(req)
	})
	req := NewRequest("I has a cat", 2, 5, []string{"have", "had", "xx"})
	scores, err := m.Score(req)
	require.NoError(t, err)
	require.Len(t, scores, 3)
	// "has" → "had"/"have" closer than "xx"
	require.Greater(t, scores[0], scores[2])

	// Java score() does not use the Guava cache — each call hits the stub.
	scores2, err := m.Score(req)
	require.NoError(t, err)
	require.Equal(t, scores, scores2)
	require.Equal(t, int32(2), atomic.LoadInt32(&scoreCalls))

	batch, err := m.BatchScore([]Request{req, NewRequest("a", 0, 1, []string{"a", "b"})})
	require.NoError(t, err)
	require.Len(t, batch, 2)
	m.Shutdown()
}

func TestRemoteLanguageModel_NilScorer(t *testing.T) {
	m := NewRemoteLanguageModel(nil)
	_, err := m.Score(NewRequest("I has a cat", 2, 5, []string{"have", "had"}))
	require.Error(t, err)
	ep := NewRemoteLanguageModelEndpoint("localhost", 5000, false, "", "", "")
	_, err = ep.Score(NewRequest("x", 0, 1, []string{"y"}))
	require.Error(t, err)
}

func TestRemoteLanguageModel_BatchScore_UsesCacheAndBatchScorer(t *testing.T) {
	var batchCalls int32
	m := NewRemoteLanguageModel(nil)
	m.BatchScorer = func(reqs []Request) ([][]float64, error) {
		atomic.AddInt32(&batchCalls, 1)
		out := make([][]float64, len(reqs))
		for i, r := range reqs {
			out[i] = make([]float64, len(r.Candidates))
			for j := range r.Candidates {
				out[i][j] = float64(j)
			}
		}
		return out, nil
	}
	r1 := NewRequest("t1", 0, 1, []string{"a", "b"})
	r2 := NewRequest("t2", 0, 1, []string{"c"})
	// first batch: both uncached → one BatchScorer call
	out, err := m.BatchScore([]Request{r1, r2})
	require.NoError(t, err)
	require.Equal(t, int32(1), atomic.LoadInt32(&batchCalls))
	require.Equal(t, []float64{0, 1}, out[0])
	// second batch: both cached → no new BatchScorer call
	out2, err := m.BatchScore([]Request{r1, r2})
	require.NoError(t, err)
	require.Equal(t, int32(1), atomic.LoadInt32(&batchCalls))
	require.Equal(t, out, out2)
	// mixed: r1 cached, r3 uncached
	r3 := NewRequest("t3", 0, 1, []string{"d"})
	_, err = m.BatchScore([]Request{r1, r3})
	require.NoError(t, err)
	require.Equal(t, int32(2), atomic.LoadInt32(&batchCalls))
}

func TestRequest_EqualHash(t *testing.T) {
	a := NewRequest("t", 1, 2, []string{"x"})
	b := NewRequest("t", 1, 2, []string{"x"})
	require.True(t, a.Equal(b))
	require.Equal(t, a.HashCode(), b.HashCode())
	c := NewRequest("t", 1, 3, []string{"x"})
	require.False(t, a.Equal(c))
}
