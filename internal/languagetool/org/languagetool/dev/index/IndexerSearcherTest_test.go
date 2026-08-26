package index

// Twin of IndexerSearcherTest — search over in-memory index.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of IndexerSearcherTest (no @Test)
func TestIndexerSearcher_NoTests(t *testing.T) {
	ix := NewIndexer()
	ix.Add("a", "the cat sat")
	ix.Add("b", "the dog ran")
	ix.Add("c", "birds fly")
	// filter + search soft
	f := NewLanguageToolFilter(true)
	toks := f.Tokenize("the cat")
	require.Equal(t, []string{"the", "cat"}, toks)
	hits := ix.SearchSubstring("cat")
	require.Equal(t, []string{"a"}, hits)
	// query builder soft
	q := NewPatternRuleQueryBuilder("f").BuildSimple(toks...)
	require.Contains(t, q, `f:"the"`)
	require.Contains(t, q, `f:"cat"`)
}
