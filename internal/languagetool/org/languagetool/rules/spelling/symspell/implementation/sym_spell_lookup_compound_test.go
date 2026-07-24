package implementation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of SymSpell.lookupCompound — Java goldens on fixture dict.
func TestLookupCompound_JavaGoldens(t *testing.T) {
	s := fixtureDict(t)

	// compound helo world => [hello world|1|5]
	got := s.LookupCompound("helo world")
	require.Equal(t, []SuggestItem{{Term: "hello world", Distance: 1, Count: 5}}, got)

	// compound helloworld => [hello world|1|5]  (split path)
	got = s.LookupCompound("helloworld")
	require.Equal(t, []SuggestItem{{Term: "hello world", Distance: 1, Count: 5}}, got)

	// compound hello => [hello|0|11]
	got = s.LookupCompound("hello")
	require.Equal(t, []SuggestItem{{Term: "hello", Distance: 0, Count: 11}}, got)
}

// Twin: single known word remains.
func TestLookupCompound_SingleKnown(t *testing.T) {
	s := DefaultSymSpell()
	require.True(t, s.CreateDictionaryEntry("test", 10, nil))
	got := s.LookupCompound("test")
	require.Equal(t, []SuggestItem{{Term: "test", Distance: 0, Count: 10}}, got)
}

// Twin: typo in first of two words.
func TestLookupCompound_Typo(t *testing.T) {
	s := DefaultSymSpell()
	require.True(t, s.CreateDictionaryEntry("hello", 100, nil))
	require.True(t, s.CreateDictionaryEntry("world", 80, nil))
	got := s.LookupCompound("helo world")
	require.Len(t, got, 1)
	require.Equal(t, "hello world", got[0].Term)
	require.Equal(t, 1, got[0].Distance)
	require.Equal(t, int64(80), got[0].Count) // min(100, 80)
}

// Twin: maxEditDistance > maxDictionaryEditDistance panics (Java IllegalArgumentException).
func TestLookupCompound_DistTooBig(t *testing.T) {
	s := DefaultSymSpell() // max edit 2
	require.PanicsWithValue(t, "Dist to big 99", func() {
		s.LookupCompoundMax("x", 99)
	})
}

// Twin: LookupCompound uses maxDictionaryEditDistance.
func TestLookupCompound_DefaultMax(t *testing.T) {
	s := NewSymSpell(16, 2, 7, 1)
	require.True(t, s.CreateDictionaryEntry("cat", 5, nil))
	require.True(t, s.CreateDictionaryEntry("dog", 5, nil))
	got := s.LookupCompound("cat dog")
	require.Equal(t, []SuggestItem{{Term: "cat dog", Distance: 0, Count: 5}}, got)
}
