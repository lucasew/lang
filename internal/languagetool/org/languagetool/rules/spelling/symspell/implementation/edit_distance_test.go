package implementation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEditDistance(t *testing.T) {
	e := NewEditDistance("hello", Damerau)
	require.Equal(t, 0, e.Compare("hello", 2))
	require.Equal(t, 1, e.Compare("hallo", 2))
	require.Equal(t, 1, e.Compare("hlelo", 2)) // transpose
	require.Equal(t, -1, e.Compare("zzzzz", 1))
	require.Equal(t, 0, NewEditDistance("", Damerau).Compare("", 1))
	require.Equal(t, 3, NewEditDistance("", Damerau).Compare("abc", 5))
}

func TestSuggestItemOrder(t *testing.T) {
	a := NewSuggestItem("a", 1, 10)
	b := NewSuggestItem("b", 1, 5)
	c := NewSuggestItem("c", 2, 100)
	require.True(t, a.Less(b))
	require.True(t, a.Less(c))
}

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
