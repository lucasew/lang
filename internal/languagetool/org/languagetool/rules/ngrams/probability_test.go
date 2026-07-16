package ngrams

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProbability(t *testing.T) {
	p := NewProbability(0.5, 0.8, 10)
	require.Equal(t, 0.5, p.GetProb())
	require.InDelta(t, math.Log(0.5), p.GetLogProb(), 1e-9)
	require.Equal(t, float32(0.8), p.GetCoverage())
	require.Equal(t, int64(10), p.GetOccurrences())
	require.Equal(t, int64(-1), NewProbabilitySimple(1, 1).GetOccurrences())
	require.Panics(t, func() { NewProbability(-0.1, 0, 0) })
}
