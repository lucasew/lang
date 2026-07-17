package index

// Twin of UnificationIndexSearchTest — soft unify-style multi-token search.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of UnificationIndexSearchTest (no @Test)
func TestUnificationIndexSearch_NoTests(t *testing.T) {
	ix := NewIndexer()
	ix.Add("1", "red car")
	ix.Add("2", "red bike")
	ix.Add("3", "blue car")
	// documents matching both "red" and "car" (soft AND via successive filter)
	red := ix.SearchSubstring("red")
	var both []string
	for _, id := range red {
		text, _ := ix.Get(id)
		if indexFold(text, "car") >= 0 {
			both = append(both, id)
		}
	}
	require.Equal(t, []string{"1"}, both)
}
