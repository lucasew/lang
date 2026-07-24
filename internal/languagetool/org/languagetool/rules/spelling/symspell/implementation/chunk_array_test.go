package implementation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Smoke twin for ChunkArray (sector ownership is separate; kept from former edit_distance_test.go).
func TestChunkArray(t *testing.T) {
	a := NewChunkArray[int](2)
	require.Equal(t, 0, a.Add(10))
	require.Equal(t, 1, a.Add(20))
	require.Equal(t, 10, a.Get(0))
	a.Set(1, 99)
	require.Equal(t, 99, a.Get(1))
	a.Clear()
	require.Equal(t, 0, a.Count)
}
