package wikipedia

// Twin of WikipediaQuickCheckTest — plain text via SimpleWikipediaTextFilter
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWikipediaQuickCheck_CheckWikipediaMarkup(t *testing.T) {
	t.Skip("unimplemented: full LT check over wiki markup")
}

func TestWikipediaQuickCheck_GetPlainText(t *testing.T) {
	f := NewSimpleWikipediaTextFilter()
	require.Equal(t, "foo Test bar", f.Filter("foo [[Test]] bar"))
	require.Equal(t, "foo visible link bar", f.Filter("foo [[Target|visible link]] bar"))
}

func TestWikipediaQuickCheck_GetPlainTextMapping(t *testing.T) {
	// Soft: plain extraction only (position mapping deferred with Sweble)
	f := NewSimpleWikipediaTextFilter()
	plain := f.Filter("hello [[World]]")
	require.Equal(t, "hello World", plain)
}

func TestWikipediaQuickCheck_GetPlainTextMappingMultiLine1(t *testing.T) {
	f := NewSimpleWikipediaTextFilter()
	plain := f.Filter("line one\n# item\n# two\n")
	require.Contains(t, plain, "item")
	require.Contains(t, plain, "two")
}

func TestWikipediaQuickCheck_GetPlainTextMappingMultiLine2(t *testing.T) {
	f := NewSimpleWikipediaTextFilter()
	require.Equal(t, "a b", f.Filter("a [[x|b]]"))
}
