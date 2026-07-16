package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringMatcherLiteralPort(t *testing.T) {
	m := NewStringMatcher("hello", false, true)
	require.True(t, m.Matches("hello"))
	require.False(t, m.Matches("Hello"))
	m2 := NewStringMatcher("hello", false, false)
	require.True(t, m2.Matches("Hello"))
}

func TestStringMatcherAlternationPort(t *testing.T) {
	m := NewStringMatcherRegexp("aa|bb")
	require.True(t, m.Matches("aa"))
	require.True(t, m.Matches("bb"))
	require.False(t, m.Matches("cc"))
	vals := m.GetPossibleValues()
	require.Contains(t, vals, "aa")
	require.Contains(t, vals, "bb")
}

func TestStringMatcherPossibleValuesPort(t *testing.T) {
	require.Nil(t, GetPossibleRegexpValues("x.*"))
	require.Nil(t, GetPossibleRegexpValues("a.c"))
	vals := GetPossibleRegexpValues("aa|bb")
	require.Contains(t, vals, "aa")
	require.Contains(t, vals, "bb")
	vals = GetPossibleRegexpValues("a(b|c)")
	require.Contains(t, vals, "ab")
	require.Contains(t, vals, "ac")
	vals = GetPossibleRegexpValues("[abc]")
	require.Len(t, vals, 3)
	vals = GetPossibleRegexpValues("are|is|w(?:as|ere)")
	require.Contains(t, vals, "was")
	require.Contains(t, vals, "were")
}

func TestStringMatcherMaxLengthPort(t *testing.T) {
	m := NewStringMatcher("x", false, true)
	long := make([]byte, MaxMatchLength+1)
	for i := range long {
		long[i] = 'x'
	}
	require.False(t, m.Matches(string(long)))
}

func TestStringMatcherInvalidSyntaxPort(t *testing.T) {
	require.Panics(t, func() { NewStringMatcherRegexp("tú|?") })
}
