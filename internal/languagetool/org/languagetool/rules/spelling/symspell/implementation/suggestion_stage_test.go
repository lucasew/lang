package implementation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Outcome twins for SuggestionStage vs Java SuggestionStage (LIFO linked list + commit merge).

func TestStageNodeAndEntryFields(t *testing.T) {
	// Node(suggestion, next)
	n := NewStageNode("word", 7)
	require.Equal(t, "word", n.Suggestion)
	require.Equal(t, 7, n.Next)

	// Entry(count, first)
	e := NewStageEntry(3, -1)
	require.Equal(t, 3, e.Count)
	require.Equal(t, -1, e.First)
}

func TestNewSuggestionStagePassesCapacityThrough(t *testing.T) {
	// Java: new HashMap<>(0); new ChunkArray<>(0) — no min-16 invent.
	s0 := NewSuggestionStage(0)
	require.NotNil(t, s0.Deletes)
	require.NotNil(t, s0.Nodes)
	require.Equal(t, 0, s0.DeleteCount())
	require.Equal(t, 0, s0.NodeCount())
	// ChunkArray(0): chunks == 0
	require.Equal(t, 0, len(s0.Nodes.Values))

	s4 := NewSuggestionStage(4)
	require.Equal(t, 0, s4.DeleteCount())
	require.Equal(t, 0, s4.NodeCount())
}

func TestSuggestionStageAddCountsAndLIFOList(t *testing.T) {
	s := NewSuggestionStage(4)
	s.Add(1, "hello")
	s.Add(1, "hallo")
	s.Add(2, "world")

	require.Equal(t, 2, s.DeleteCount())
	require.Equal(t, 3, s.NodeCount())

	// Entry for key 1: two nodes, first points at most recently added ("hallo")
	e1 := s.Deletes[1]
	require.Equal(t, 2, e1.Count)
	require.Equal(t, 1, e1.First) // index of "hallo" node
	nHallo := s.Nodes.Get(e1.First)
	require.Equal(t, "hallo", nHallo.Suggestion)
	require.Equal(t, 0, nHallo.Next) // → "hello"
	nHello := s.Nodes.Get(nHallo.Next)
	require.Equal(t, "hello", nHello.Suggestion)
	require.Equal(t, -1, nHello.Next)

	e2 := s.Deletes[2]
	require.Equal(t, 1, e2.Count)
	require.Equal(t, 2, e2.First)
	nWorld := s.Nodes.Get(e2.First)
	require.Equal(t, "world", nWorld.Suggestion)
	require.Equal(t, -1, nWorld.Next)
}

func TestSuggestionStageCommitToLIFOOrder(t *testing.T) {
	s := NewSuggestionStage(4)
	s.Add(1, "hello")
	s.Add(1, "hallo")
	s.Add(2, "world")

	perm := map[int][]string{}
	s.CommitTo(perm)

	// Walk from first yields most recently added first → staged append order is LIFO.
	require.Equal(t, []string{"hallo", "hello"}, perm[1])
	require.Equal(t, []string{"world"}, perm[2])
}

func TestSuggestionStageCommitToMergeExisting(t *testing.T) {
	// Java: if permanentDeletes.containsKey(key), Arrays.copyOf grow; existing stay prefix.
	s := NewSuggestionStage(4)
	s.Add(1, "a")
	s.Add(1, "b")
	s.Add(1, "c")

	perm := map[int][]string{
		1: {"existing1", "existing2"},
		9: {"untouched"},
	}
	s.CommitTo(perm)

	// existing prefix + LIFO staged (c, then b, then a)
	require.Equal(t, []string{"existing1", "existing2", "c", "b", "a"}, perm[1])
	require.Equal(t, []string{"untouched"}, perm[9])
}

func TestSuggestionStageCommitToMultiKey(t *testing.T) {
	s := NewSuggestionStage(8)
	s.Add(10, "x1")
	s.Add(20, "y1")
	s.Add(10, "x2")
	s.Add(30, "z1")
	s.Add(20, "y2")

	perm := map[int][]string{}
	s.CommitTo(perm)

	require.Equal(t, []string{"x2", "x1"}, perm[10])
	require.Equal(t, []string{"y2", "y1"}, perm[20])
	require.Equal(t, []string{"z1"}, perm[30])
	require.Equal(t, 3, s.DeleteCount())
	require.Equal(t, 5, s.NodeCount())
}

func TestSuggestionStageClear(t *testing.T) {
	s := NewSuggestionStage(4)
	s.Add(1, "hello")
	s.Add(2, "world")
	require.Equal(t, 2, s.DeleteCount())
	require.Equal(t, 2, s.NodeCount())

	s.Clear()
	require.Equal(t, 0, s.DeleteCount())
	require.Equal(t, 0, s.NodeCount())

	// After clear, can stage again and commit cleanly.
	s.Add(3, "again")
	perm := map[int][]string{}
	s.CommitTo(perm)
	require.Equal(t, []string{"again"}, perm[3])
	require.Equal(t, 1, s.DeleteCount())
	require.Equal(t, 1, s.NodeCount())
}

func TestSuggestionStageCommitAfterClearDoesNotReplayOld(t *testing.T) {
	s := NewSuggestionStage(4)
	s.Add(1, "old")
	s.Clear()
	s.Add(1, "new")

	perm := map[int][]string{}
	s.CommitTo(perm)
	require.Equal(t, []string{"new"}, perm[1])
}
