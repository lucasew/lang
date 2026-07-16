package implementation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSymSpellExactAndTypo(t *testing.T) {
	s := DefaultSymSpell()
	require.True(t, s.CreateDictionaryEntry("hello", 10, nil))
	require.True(t, s.CreateDictionaryEntry("world", 5, nil))
	require.False(t, s.CreateDictionaryEntry("hello", 1, nil)) // update only
	require.Equal(t, 2, s.WordCount())

	// exact
	got := s.Lookup("hello", VerbosityTop)
	require.Len(t, got, 1)
	require.Equal(t, "hello", got[0].Term)
	require.Equal(t, 0, got[0].Distance)

	// single edit (swap / substitution-ish delete path)
	got = s.Lookup("helo", VerbosityClosest)
	require.NotEmpty(t, got)
	require.Equal(t, "hello", got[0].Term)
	require.LessOrEqual(t, got[0].Distance, 2)
}

func TestSymSpellStaging(t *testing.T) {
	s := DefaultSymSpell()
	st := NewSuggestionStage(16)
	require.True(t, s.CreateDictionaryEntry("test", 3, st))
	require.True(t, s.CreateDictionaryEntry("toast", 2, st))
	s.CommitStaging(st)
	got := s.Lookup("test", VerbosityTop)
	require.Len(t, got, 1)
	require.Equal(t, "test", got[0].Term)
}

func TestSymSpellBelowThreshold(t *testing.T) {
	s := NewSymSpell(16, 2, 7, 3)
	require.False(t, s.CreateDictionaryEntry("rare", 1, nil))
	require.Equal(t, 0, s.WordCount())
	require.False(t, s.CreateDictionaryEntry("rare", 1, nil))
	require.True(t, s.CreateDictionaryEntry("rare", 1, nil)) // reaches 3
	require.Equal(t, 1, s.WordCount())
}
