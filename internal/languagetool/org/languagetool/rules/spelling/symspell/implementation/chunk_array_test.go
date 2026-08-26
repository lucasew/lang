package implementation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Outcome twins for ChunkArray vs Java ChunkArray (not smoke-only).

func TestChunkArray_InitialCapacityZero(t *testing.T) {
	// Java: chunks = (0 + 4096 - 1) / 4096 = 0; values = new Node[0][]
	a := NewChunkArray[int](0)
	require.Equal(t, 0, len(a.Values), "initialCapacity=0 must allocate 0 chunks (no invent min-1)")
	require.Equal(t, 0, a.capacity())
	require.Equal(t, 0, a.Count)

	// First add grows by one chunk of 4096
	idx := a.Add(42)
	require.Equal(t, 0, idx)
	require.Equal(t, 1, len(a.Values))
	require.Equal(t, chunkArrayChunkSize, len(a.Values[0]))
	require.Equal(t, chunkArrayChunkSize, a.capacity())
	require.Equal(t, 1, a.Count)
	require.Equal(t, 42, a.Get(0))
}

func TestChunkArray_InitialCapacityChunkFormula(t *testing.T) {
	// chunks = (initialCapacity + chunkSize - 1) / chunkSize
	cases := []struct {
		cap    int
		chunks int
	}{
		{1, 1},
		{4095, 1},
		{4096, 1},
		{4097, 2},
		{8192, 2},
		{8193, 3},
	}
	for _, tc := range cases {
		a := NewChunkArray[int](tc.cap)
		require.Equal(t, tc.chunks, len(a.Values), "initialCapacity=%d", tc.cap)
		require.Equal(t, tc.chunks*chunkArrayChunkSize, a.capacity(), "initialCapacity=%d", tc.cap)
		for i := range a.Values {
			require.Equal(t, chunkArrayChunkSize, len(a.Values[i]), "chunk %d width", i)
		}
		require.Equal(t, 0, a.Count)
	}
}

func TestChunkArray_AddGetSetRoundtrip(t *testing.T) {
	a := NewChunkArray[int](2)
	require.Equal(t, 0, a.Add(10))
	require.Equal(t, 1, a.Add(20))
	require.Equal(t, 2, a.Add(30))
	require.Equal(t, 10, a.Get(0))
	require.Equal(t, 20, a.Get(1))
	require.Equal(t, 30, a.Get(2))
	a.Set(1, 99)
	require.Equal(t, 99, a.Get(1))
	require.Equal(t, 10, a.Get(0))
	require.Equal(t, 30, a.Get(2))
	require.Equal(t, 3, a.Count)
}

func TestChunkArray_GrowPastFirstChunkKeepsPrior(t *testing.T) {
	// Start with capacity for exactly one chunk; fill and exceed.
	a := NewChunkArray[int](1)
	require.Equal(t, 1, len(a.Values))
	require.Equal(t, chunkArrayChunkSize, a.capacity())

	for i := 0; i < chunkArrayChunkSize; i++ {
		idx := a.Add(i)
		require.Equal(t, i, idx)
	}
	require.Equal(t, chunkArrayChunkSize, a.Count)
	require.Equal(t, 1, len(a.Values), "still one chunk when count == capacity")

	// Next add must grow outer array by +1 chunk
	idx := a.Add(999_001)
	require.Equal(t, chunkArrayChunkSize, idx)
	require.Equal(t, 2, len(a.Values))
	require.Equal(t, 2*chunkArrayChunkSize, a.capacity())
	require.Equal(t, chunkArrayChunkSize+1, a.Count)

	// Prior values intact across growth
	require.Equal(t, 0, a.Get(0))
	require.Equal(t, chunkArrayChunkSize-1, a.Get(chunkArrayChunkSize-1))
	require.Equal(t, 999_001, a.Get(chunkArrayChunkSize))
}

func TestChunkArray_ClearOnlyResetsCount(t *testing.T) {
	a := NewChunkArray[int](8)
	a.Add(11)
	a.Add(22)
	require.Equal(t, 2, a.Count)
	capBefore := a.capacity()
	chunksBefore := len(a.Values)

	a.Clear()
	require.Equal(t, 0, a.Count)
	// Java clear only sets count=0; does not shrink
	require.Equal(t, capBefore, a.capacity())
	require.Equal(t, chunksBefore, len(a.Values))
	// Residual storage left (Java leaves arrays); Get of old indices still returns prior values
	require.Equal(t, 11, a.Get(0))
	require.Equal(t, 22, a.Get(1))

	// After clear, add reuses from index 0
	require.Equal(t, 0, a.Add(77))
	require.Equal(t, 77, a.Get(0))
	require.Equal(t, 1, a.Count)
}

func TestChunkArray_RowColAtChunkBoundaries(t *testing.T) {
	// Need at least 2 chunks of storage for indices 4095 and 4096
	a := NewChunkArray[int](chunkArrayChunkSize + 1)
	require.Equal(t, 2, len(a.Values))

	// Fill through first chunk boundary and one into second
	for i := 0; i <= chunkArrayChunkSize; i++ {
		a.Add(1000 + i)
	}
	// index 4095 → row 0, col 4095
	require.Equal(t, 0, a.row(4095))
	require.Equal(t, 4095, a.col(4095))
	require.Equal(t, 1000+4095, a.Get(4095))
	// index 4096 → row 1, col 0
	require.Equal(t, 1, a.row(4096))
	require.Equal(t, 0, a.col(4096))
	require.Equal(t, 1000+4096, a.Get(4096))
}

func TestChunkArray_SetOnExternalList(t *testing.T) {
	// Twin of setValues(index, value, list)
	a := NewChunkArray[int](1)
	list := make([][]int, 1)
	list[0] = make([]int, chunkArrayChunkSize)
	a.SetOn(list, 3, 55)
	require.Equal(t, 55, list[0][3])
}
