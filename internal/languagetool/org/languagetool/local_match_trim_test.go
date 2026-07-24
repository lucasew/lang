package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalMatch_TrimMatchEnds_TrailingToken(t *testing.T) {
	// Java: error "foo bar" + all suggestions end with " bar" → trim to "foo" / "baz"
	m := LocalMatch{
		FromPos: 0, ToPos: 7,
		OriginalErrorStr: "foo bar",
		Suggestions:      []string{"baz bar", "qux bar"},
		FromPosSentence:  0, ToPosSentence: 7,
	}
	out := m.TrimMatchEnds()
	require.Equal(t, 0, out.FromPos)
	require.Equal(t, 3, out.ToPos) // "foo"
	require.Equal(t, "foo", out.OriginalErrorStr)
	require.Equal(t, []string{"baz", "qux"}, out.Suggestions)
	require.Equal(t, 3, out.ToPosSentence)
}

func TestLocalMatch_TrimMatchEnds_LeadingToken(t *testing.T) {
	m := LocalMatch{
		FromPos: 5, ToPos: 12,
		OriginalErrorStr: "the word",
		Suggestions:      []string{"the fix", "the ok"},
		FromPosSentence:  5, ToPosSentence: 12,
	}
	out := m.TrimMatchEnds()
	require.Equal(t, 9, out.FromPos) // +4 for "the "
	require.Equal(t, 12, out.ToPos)
	require.Equal(t, "word", out.OriginalErrorStr)
	require.Equal(t, []string{"fix", "ok"}, out.Suggestions)
}

func TestLocalMatch_TrimMatchEnds_NoChangeWhenMismatch(t *testing.T) {
	m := LocalMatch{
		FromPos: 0, ToPos: 7,
		OriginalErrorStr: "foo bar",
		Suggestions:      []string{"baz bar", "other"},
	}
	out := m.TrimMatchEnds()
	require.Equal(t, m.FromPos, out.FromPos)
	require.Equal(t, m.ToPos, out.ToPos)
	require.Equal(t, m.Suggestions, out.Suggestions)
}

func TestLocalMatch_TrimMatchEnds_EmptySurface(t *testing.T) {
	m := LocalMatch{FromPos: 0, ToPos: 3, Suggestions: []string{"x y"}}
	out := m.TrimMatchEnds()
	require.Equal(t, m, out)
}

func TestLocalMatch_TrimMatchEnds_UTF16NonBMP(t *testing.T) {
	// Java String.length counts "😀" as 2 UTF-16 units; space+emoji = 3.
	// error "a 😀" (UTF-16 len 4: a, space, hi, lo) → trim trailing " 😀" → "a" (toPos 1)
	emoji := "😀"
	errorStr := "a " + emoji
	// positions as Java UTF-16: 0..len
	utf16To := 0
	for _, r := range errorStr {
		if r >= 0x10000 {
			utf16To += 2
		} else {
			utf16To++
		}
	}
	m := LocalMatch{
		FromPos: 0, ToPos: utf16To,
		OriginalErrorStr: errorStr,
		Suggestions:      []string{"b " + emoji, "c " + emoji},
		FromPosSentence:  0, ToPosSentence: utf16To,
	}
	out := m.TrimMatchEnds()
	require.Equal(t, "a", out.OriginalErrorStr)
	require.Equal(t, []string{"b", "c"}, out.Suggestions)
	require.Equal(t, 0, out.FromPos)
	require.Equal(t, 1, out.ToPos) // "a" is one UTF-16 unit
	require.Equal(t, 1, out.ToPosSentence)
}

func TestLocalMatch_TrimMatchEnds_UTF16Leading(t *testing.T) {
	// "😀 x" trim leading emoji token → "x"; fromPos advances by UTF-16 len of "😀 "
	emoji := "😀"
	errorStr := emoji + " x"
	utf16Prefix := 0
	for _, r := range emoji + " " {
		if r >= 0x10000 {
			utf16Prefix += 2
		} else {
			utf16Prefix++
		}
	}
	utf16All := 0
	for _, r := range errorStr {
		if r >= 0x10000 {
			utf16All += 2
		} else {
			utf16All++
		}
	}
	m := LocalMatch{
		FromPos: 10, ToPos: 10 + utf16All,
		OriginalErrorStr: errorStr,
		Suggestions:      []string{emoji + " y"},
	}
	out := m.TrimMatchEnds()
	require.Equal(t, "x", out.OriginalErrorStr)
	require.Equal(t, []string{"y"}, out.Suggestions)
	require.Equal(t, 10+utf16Prefix, out.FromPos)
}
