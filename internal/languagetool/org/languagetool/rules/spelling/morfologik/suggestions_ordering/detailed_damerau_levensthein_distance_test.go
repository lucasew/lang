package suggestions_ordering

import (
	"testing"
	"unicode/utf16"

	"github.com/stretchr/testify/require"
)

func TestDistanceValueAndMutators(t *testing.T) {
	d := Distance{}
	require.Equal(t, 0, d.Value())
	require.Equal(t, Distance{Inserts: 1}, d.Insert())
	require.Equal(t, Distance{Deletes: 1}, d.Delete())
	require.Equal(t, Distance{Replaces: 1}, d.Replace())
	require.Equal(t, Distance{Transposes: 1}, d.Transpose())
	require.Equal(t, 4, d.Insert().Delete().Replace().Transpose().Value())
}

func TestDistanceTrack(t *testing.T) {
	d := Distance{}
	require.Equal(t, Distance{Inserts: 1}, d.Track(NewInsert()))
	require.Equal(t, Distance{Deletes: 1}, d.Track(NewDelete()))
	require.Equal(t, Distance{Replaces: 1}, d.Track(NewReplace()))
	require.Equal(t, Distance{Transposes: 1}, d.Track(NewTranspose()))
	require.Panics(t, func() {
		type unknownOp struct{}
		// force unknown via interface with a dummy that implements Apply
		d.Track(bogusOp{})
	})
}

type bogusOp struct{}

func (bogusOp) Apply(string) (string, bool) { return "", false }

func TestDistanceEqualsAndString(t *testing.T) {
	a := NewDistance(1, 2, 3, 4)
	b := NewDistance(1, 2, 3, 4)
	c := NewDistance(0, 2, 3, 4)
	require.True(t, a.Equals(b))
	require.False(t, a.Equals(c))
	require.Contains(t, a.String(), "value=10")
	require.Contains(t, a.String(), "inserts=1")
	require.Contains(t, a.String(), "deletes=2")
	require.Contains(t, a.String(), "replaces=3")
	require.Contains(t, a.String(), "transposes=4")
}

func TestCompareIdentical(t *testing.T) {
	require.Equal(t, Distance{}, Compare("hello", "hello"))
	require.Equal(t, 0, Compare("", "").Value())
}

func TestCompareSingleOps(t *testing.T) {
	// single replace
	d := Compare("cat", "bat")
	require.Equal(t, Distance{Replaces: 1}, d)
	// transpose
	require.Equal(t, Distance{Transposes: 1}, Compare("ab", "ba"))
	// insert
	require.Equal(t, Distance{Inserts: 1}, Compare("cat", "cats"))
	// delete
	require.Equal(t, Distance{Deletes: 1}, Compare("cats", "cat"))
}

func TestCompareUTF16Units(t *testing.T) {
	// Java charAt: emoji is two UTF-16 code units.
	// JVM: compare("a😀","aX") → value=2 deletes=1 replaces=1
	d := Compare("a😀", "aX")
	require.Equal(t, Distance{Deletes: 1, Replaces: 1}, d)
	require.Equal(t, 2, d.Value())
	// BMP accented letter: one code unit each side → single replace
	require.Equal(t, Distance{Replaces: 1}, Compare("café", "cafe"))
}

func TestEditOperationDelete(t *testing.T) {
	del := NewDelete()
	_, ok := del.Apply("")
	require.False(t, ok) // Java null
	_, ok = del.Apply("a")
	require.False(t, ok)

	// length decreases by one UTF-16 unit when applied
	s, ok := NewDeleteSeeded(1).Apply("ab")
	require.True(t, ok)
	require.Equal(t, 1, len(utf16.Encode([]rune(s))))
}

func TestEditOperationInsert(t *testing.T) {
	s, ok := NewInsertSeeded(0).Apply("hi")
	require.True(t, ok)
	require.Equal(t, 3, len(utf16.Encode([]rune(s))))
	// empty source: always inserts one lowercase letter
	s, ok = NewInsertSeeded(0).Apply("")
	require.True(t, ok)
	require.Equal(t, 1, len(utf16.Encode([]rune(s))))
	r := []rune(s)
	require.GreaterOrEqual(t, r[0], 'a')
	require.LessOrEqual(t, r[0], 'z')
}

func TestEditOperationReplace(t *testing.T) {
	_, ok := NewReplace().Apply("")
	require.False(t, ok)
	s, ok := NewReplaceSeeded(1).Apply("ab")
	require.True(t, ok)
	require.Equal(t, 2, len(utf16.Encode([]rune(s))))
}

func TestEditOperationTranspose(t *testing.T) {
	_, ok := NewTranspose().Apply("a")
	require.False(t, ok)
	_, ok = NewTranspose().Apply("")
	require.False(t, ok)
	s, ok := NewTransposeSeeded(1).Apply("ab")
	require.True(t, ok)
	// adjacent swap of two-char string is the reverse
	require.Equal(t, "ba", s)
	s, ok = NewTransposeSeeded(5).Apply("abc")
	require.True(t, ok)
	require.Equal(t, 3, len(s))
	// result is a permutation with exactly one adjacent transposition from "abc"
	require.Contains(t, []string{"bac", "acb"}, s)
}

func TestRandomEdit(t *testing.T) {
	require.Len(t, EditOperations, 4)
	op := RandomEdit()
	require.NotNil(t, op)
	// can apply to a long enough string without panic
	_, _ = op.Apply("abcdef")
}
