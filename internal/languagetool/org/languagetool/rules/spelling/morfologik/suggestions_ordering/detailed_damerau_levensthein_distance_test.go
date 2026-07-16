package suggestions_ordering

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompareIdentical(t *testing.T) {
	d := Compare("hello", "hello")
	require.Equal(t, 0, d.Value())
}

func TestCompareOps(t *testing.T) {
	// single replace
	d := Compare("cat", "bat")
	require.Equal(t, 1, d.Value())
	require.Equal(t, 1, d.Replaces)

	// transpose
	d = Compare("ab", "ba")
	require.Equal(t, 1, d.Value())
	require.Equal(t, 1, d.Transposes)

	// insert
	d = Compare("cat", "cats")
	require.Equal(t, 1, d.Value())
	require.Equal(t, 1, d.Inserts)

	// delete
	d = Compare("cats", "cat")
	require.Equal(t, 1, d.Value())
	require.Equal(t, 1, d.Deletes)
}

func TestEditOperations(t *testing.T) {
	del := &DeleteOp{}
	del.r = nil
	// with seeded default
	s := (&InsertOp{}).Apply("hi")
	require.Len(t, []rune(s), 3)
}
