package index

// Twin of LanguageToolFilterTest — soft tokenize for indexing.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of LanguageToolFilterTest (no @Test)
func TestLanguageToolFilter_NoTests(t *testing.T) {
	f := NewLanguageToolFilter(true)
	toks := f.Tokenize("Hello, World! 123")
	require.Equal(t, []string{"hello", "world", "123"}, toks)

	f2 := NewLanguageToolFilter(false)
	require.Equal(t, []string{"Hello", "World"}, f2.Tokenize("Hello World"))
}
