package implementation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuggestionStage(t *testing.T) {
	s := NewSuggestionStage(4)
	s.Add(1, "hello")
	s.Add(1, "hallo")
	s.Add(2, "world")
	require.Equal(t, 2, s.DeleteCount())
	require.Equal(t, 3, s.NodeCount())
	perm := map[int][]string{}
	s.CommitTo(perm)
	require.ElementsMatch(t, []string{"hello", "hallo"}, perm[1])
	require.Equal(t, []string{"world"}, perm[2])
	s.Clear()
	require.Equal(t, 0, s.DeleteCount())
}
