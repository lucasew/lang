package index

// Twin of IndexerTest — in-memory index smoke (Lucene deferred).
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of IndexerTest (no @Test)
func TestIndexer_NoTests(t *testing.T) {
	ix := NewIndexer()
	ix.Add("1", "hello world")
	ix.Add("2", "goodbye moon")
	require.Equal(t, 2, ix.Size())
	text, ok := ix.Get("1")
	require.True(t, ok)
	require.Equal(t, "hello world", text)
	hits := ix.SearchSubstring("world")
	require.Contains(t, hits, "1")
	require.NotContains(t, hits, "2")
}
