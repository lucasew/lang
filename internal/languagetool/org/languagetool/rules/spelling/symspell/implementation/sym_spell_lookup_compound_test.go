package implementation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of SymSpell.lookupCompound — join two terms that form a known word.
func TestLookupCompound_Combi(t *testing.T) {
	s := DefaultSymSpell()
	// "helloworld" as one word; also "hello" and "world"
	require.True(t, s.CreateDictionaryEntry("hello", 100, nil))
	require.True(t, s.CreateDictionaryEntry("world", 80, nil))
	require.True(t, s.CreateDictionaryEntry("helloworld", 50, nil))

	// input "hello world" (space between known words) — compound path may join
	got := s.LookupCompound("hello world")
	require.Len(t, got, 1)
	require.NotEmpty(t, got[0].Term)

	// "helloworld" as single term without space — split path may recover "hello world"
	got = s.LookupCompound("helloworld")
	require.Len(t, got, 1)
	// Prefer split into hello world or exact helloworld
	require.True(t,
		got[0].Term == "helloworld" || got[0].Term == "hello world",
		"got %q", got[0].Term)
}

// Twin: single known word remains.
func TestLookupCompound_SingleKnown(t *testing.T) {
	s := DefaultSymSpell()
	require.True(t, s.CreateDictionaryEntry("test", 10, nil))
	got := s.LookupCompound("test")
	require.Len(t, got, 1)
	require.Equal(t, "test", got[0].Term)
	require.Equal(t, 0, got[0].Distance)
}

// Twin: typo in first of two words.
func TestLookupCompound_Typo(t *testing.T) {
	s := DefaultSymSpell()
	require.True(t, s.CreateDictionaryEntry("hello", 100, nil))
	require.True(t, s.CreateDictionaryEntry("world", 80, nil))
	got := s.LookupCompound("helo world")
	require.Len(t, got, 1)
	// Should correct helo → hello
	require.Contains(t, got[0].Term, "hello")
	require.Contains(t, got[0].Term, "world")
}

// Twin: maxEditDistance > maxDictionaryEditDistance panics (Java IllegalArgumentException).
func TestLookupCompound_DistTooBig(t *testing.T) {
	s := DefaultSymSpell() // max edit 2
	require.Panics(t, func() {
		s.LookupCompoundMax("x", 99)
	})
}

// Twin: LookupCompound uses maxDictionaryEditDistance.
func TestLookupCompound_DefaultMax(t *testing.T) {
	s := NewSymSpell(16, 2, 7, 1)
	require.True(t, s.CreateDictionaryEntry("cat", 5, nil))
	require.True(t, s.CreateDictionaryEntry("dog", 5, nil))
	got := s.LookupCompound("cat dog")
	require.Len(t, got, 1)
	require.Equal(t, "cat dog", got[0].Term)
}
