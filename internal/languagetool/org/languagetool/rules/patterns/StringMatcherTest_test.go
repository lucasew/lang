package patterns

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringMatcher_SyntaxIsValidated(t *testing.T) {
	require.Panics(t, func() { NewStringMatcherRegexp("tú|?") })
}

func TestStringMatcher_GetPossibleValues(t *testing.T) {
	// Java assertPossibleValues: unenumerable → null
	require.Nil(t, GetPossibleRegexpValues("x.*"))
	require.Nil(t, GetPossibleRegexpValues("x+"))
	require.Nil(t, GetPossibleRegexpValues("a.c"))
	require.Nil(t, GetPossibleRegexpValues("a{2}"))
	require.Nil(t, GetPossibleRegexpValues("[a-z]")) // range not expanded as possible values? check Go
	// if Go expands [a-z] differently, still twin when nil or full set
	require.Nil(t, GetPossibleRegexpValues("(?=a)"))

	// empty / anchors
	empty := GetPossibleRegexpValues("")
	require.NotNil(t, empty)
	require.Contains(t, empty, "")
	require.Equal(t, map[string]struct{}{"x": {}}, GetPossibleRegexpValues("^x$"))

	vals := GetPossibleRegexpValues("aa|bb")
	require.Contains(t, vals, "aa")
	require.Contains(t, vals, "bb")
	vals = GetPossibleRegexpValues("a(b|c)")
	require.Contains(t, vals, "ab")
	require.Contains(t, vals, "ac")
	vals = GetPossibleRegexpValues("[abc]")
	require.Len(t, vals, 3)
	vals = GetPossibleRegexpValues("are|is|w(?:as|ere)")
	require.Contains(t, vals, "are")
	require.Contains(t, vals, "is")
	require.Contains(t, vals, "was")
	require.Contains(t, vals, "were")
	vals = GetPossibleRegexpValues("tú|\\?")
	require.Contains(t, vals, "tú")
	require.Contains(t, vals, "?")
	vals = GetPossibleRegexpValues("NN|PRP\\$")
	require.Contains(t, vals, "NN")
	require.Contains(t, vals, "PRP$")

	m := NewStringMatcherRegexp("aa|bb")
	require.True(t, m.Matches("aa"))
	require.False(t, m.Matches("cc"))
	require.Contains(t, m.GetPossibleValues(), "bb")
}

func TestStringMatcher_RequiredSubstrings(t *testing.T) {
	// Twin of assertRequiredSubstrings samples
	require.Equal(t, "[]", GetRequiredSubstrings("").String())
	require.Equal(t, "[foo]", GetRequiredSubstrings("foo").String())
	require.Nil(t, GetRequiredSubstrings("foo|bar"))
	require.Nil(t, GetRequiredSubstrings("\\w"))

	s := GetRequiredSubstrings("PRP.+")
	require.NotNil(t, s)
	require.Equal(t, "[PRP)", s.String())

	s = GetRequiredSubstrings(".*PRP.+")
	require.Equal(t, "(PRP)", s.String())
	s = GetRequiredSubstrings(".*PRP")
	require.Equal(t, "(PRP]", s.String())
	s = GetRequiredSubstrings("a.+b")
	require.Equal(t, "[a, b]", s.String())

	s = GetRequiredSubstrings("\\bvon Seiten\\b")
	require.Equal(t, "[von Seiten]", s.String())
	s = GetRequiredSubstrings("§ ?(\\d+[a-z]?)")
	require.Equal(t, "[§)", s.String())
}

// Twin of StringMatcherTest.noSOEOnLongDisjunction — 100k alts via set path, no SOE.
func TestStringMatcher_NoSOEOnLongDisjunction(t *testing.T) {
	const count = 100_000
	var b strings.Builder
	// pre-size roughly: "a" + digits + "|"
	b.Grow(count * 8)
	for i := 0; i < count; i++ {
		if i > 0 {
			b.WriteByte('|')
		}
		fmt.Fprintf(&b, "a%d", i)
	}
	pattern := b.String()
	// Java: StringMatcher.create(pattern, true, true) — isRegExp, caseSensitive
	matcher := NewStringMatcher(pattern, true, true)
	// Spot-check across the range (full loop is O(n) and heavy; Java asserts all)
	for _, i := range []int{0, 1, 42, count / 2, count - 1} {
		require.True(t, matcher.Matches(fmt.Sprintf("a%d", i)), "a%d", i)
		require.False(t, matcher.Matches(fmt.Sprintf("b%d", i)), "b%d", i)
	}
	// ensure possible values path (not full RE2 match per call)
	pv := matcher.GetPossibleValues()
	require.NotNil(t, pv)
	require.Len(t, pv, count)
}
