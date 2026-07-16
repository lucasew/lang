package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringMatcher_SyntaxIsValidated(t *testing.T) {
	require.Panics(t, func() { NewStringMatcherRegexp("tú|?") })
}

func TestStringMatcher_GetPossibleValues(t *testing.T) {
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
	m := NewStringMatcherRegexp("aa|bb")
	require.True(t, m.Matches("aa"))
	require.Contains(t, m.GetPossibleValues(), "bb")
}

func TestStringMatcher_RequiredSubstrings(t *testing.T) {
	s := NewSubstrings(false, false, []string{"foo"})
	pos := s.Find("xxfoozz", true)
	require.GreaterOrEqual(t, pos, 0)
	// missing fragment
	require.Equal(t, -1, s.Find("xxbarzz", true))
	s2 := NewSubstrings(true, true, []string{"hello"})
	if refined := s2.CheckCanReplaceRegex("hello"); refined != nil {
		require.True(t, refined.MustStart)
	}
}
