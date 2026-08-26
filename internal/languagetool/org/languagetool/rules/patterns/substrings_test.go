package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubstrings_Matches(t *testing.T) {
	s := NewSubstrings(true, true, []string{"ab", "cd"})
	require.True(t, s.Matches("abXXcd", true))
	require.False(t, s.Matches("XabXXcd", true))
	require.False(t, s.Matches("abXXc", true))

	s2 := NewSubstrings(false, false, []string{"foo"})
	require.True(t, s2.Matches("xxfooyy", true))
	require.Equal(t, 2, s2.Find("xxfooyy", true))

	// concat merge when mustEnd+mustStart
	a := NewSubstrings(true, true, []string{"pre"})
	b := NewSubstrings(true, false, []string{"fix"})
	c := a.Concat(b)
	require.Equal(t, []string{"prefix"}, c.Substrings)
	require.True(t, c.MustStart)
	require.False(t, c.MustEnd)

	require.NotNil(t, NewSubstrings(true, true, []string{"a", "b"}).CheckCanReplaceRegex("a.*b"))
}

func TestSubstrings_IgnoreCase(t *testing.T) {
	s := NewSubstrings(false, false, []string{"Foo"})
	require.True(t, s.Matches("xxfooyy", false))
}
