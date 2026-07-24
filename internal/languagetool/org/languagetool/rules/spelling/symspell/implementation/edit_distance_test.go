package implementation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin outcomes for org.languagetool.rules.spelling.symspell.implementation.EditDistance
// (and nested DistanceAlgorithm). Empty Go string maps Java null/empty per edit_distance.go.

func TestEditDistance_IdentityInsertDeleteSubstituteTranspose(t *testing.T) {
	e := NewEditDistance("hello", Damerau)
	require.Equal(t, 0, e.Compare("hello", 2))          // identity
	require.Equal(t, 1, e.Compare("hallo", 2))          // substitute e→a
	require.Equal(t, 1, e.Compare("hlelo", 2))          // transpose el→le
	require.Equal(t, 1, e.Compare("helo", 2))           // delete l
	require.Equal(t, 1, e.Compare("helllo", 2))         // insert l
	require.Equal(t, 1, e.Compare("hxllo", 1))          // substitute within budget
	require.Equal(t, 1, NewEditDistance("ab", Damerau).Compare("ba", 2)) // pure transpose
	require.Equal(t, 1, NewEditDistance("a", Damerau).Compare("b", 1))
	require.Equal(t, 0, NewEditDistance("a", Damerau).Compare("a", 1))
	require.Equal(t, 2, NewEditDistance("book", Damerau).Compare("back", 2))
	require.Equal(t, 1, NewEditDistance("abcdef", Damerau).Compare("abcxef", 1))
}

func TestEditDistance_MaxDistanceZero(t *testing.T) {
	// Equal → 0; unequal → -1. Empty base is null path and skips maxDistance==0 branch.
	require.Equal(t, 0, NewEditDistance("abc", Damerau).Compare("abc", 0))
	require.Equal(t, -1, NewEditDistance("abc", Damerau).Compare("abd", 0))
	require.Equal(t, -1, NewEditDistance("abc", Damerau).Compare("xyz", 0))
	require.Equal(t, -1, NewEditDistance("abcdef", Damerau).Compare("abcxef", 0))
	// Java null base returns string2.length() before maxDistance check
	require.Equal(t, 3, NewEditDistance("", Damerau).Compare("abc", 0))
	require.Equal(t, 0, NewEditDistance("", Damerau).Compare("", 0))
	// non-null base + empty string2 returns base.length() before maxDistance check
	require.Equal(t, 3, NewEditDistance("abc", Damerau).Compare("", 0))
}

func TestEditDistance_EmptyBaseAndReverse(t *testing.T) {
	// Java empty ctor → baseString=null
	empty := NewEditDistance("", Damerau)
	require.Equal(t, 0, empty.Compare("", 1))
	require.Equal(t, 3, empty.Compare("abc", 5))
	require.Equal(t, 5, empty.Compare("hello", 10))

	// non-empty base vs empty string2
	require.Equal(t, 3, NewEditDistance("abc", Damerau).Compare("", 5))
	require.Equal(t, 5, NewEditDistance("hello", Damerau).Compare("", 10))
	// length-only path ignores maxDistance budget
	require.Equal(t, 5, NewEditDistance("hello", Damerau).Compare("", 0))
	require.Equal(t, 5, NewEditDistance("hello", Damerau).Compare("", 2))
}

func TestEditDistance_MaxDistanceExceeded(t *testing.T) {
	e := NewEditDistance("hello", Damerau)
	require.Equal(t, -1, e.Compare("zzzzz", 1))
	require.Equal(t, 5, e.Compare("zzzzz", 5))
	require.Equal(t, -1, NewEditDistance("kitten", Damerau).Compare("sitting", 2))
	require.Equal(t, 3, NewEditDistance("kitten", Damerau).Compare("sitting", 3))
	require.Equal(t, -1, NewEditDistance("xy", Damerau).Compare("ab", 1))
	require.Equal(t, -1, NewEditDistance("short", Damerau).Compare("longerword", 3))
	// Pure prefix/suffix path returns tLen before maxDistance clamp (Java early return).
	require.Equal(t, 2, NewEditDistance("abcd", Damerau).Compare("ab", 1))
	require.Equal(t, 2, NewEditDistance("abcd", Damerau).Compare("ab", 5))
	require.Equal(t, 2, NewEditDistance("ab", Damerau).Compare("abcd", 5))
	// lenDiff > maxDistance when remaining after affix strip still too long
	require.Equal(t, -1, NewEditDistance("axyz", Damerau).Compare("b", 1))
	// ac → abxy distance 3; budgets 1 and 2 both exceed
	require.Equal(t, -1, NewEditDistance("ac", Damerau).Compare("abxy", 2))
	require.Equal(t, 3, NewEditDistance("ac", Damerau).Compare("abxy", 3))
}

func TestEditDistance_UTF16CodeUnits(t *testing.T) {
	// café: c,a,f,é — 4 BMP UTF-16 units (é is U+00E9, not combining)
	require.Equal(t, 4, NewEditDistance("", Damerau).Compare("café", 10))
	require.Equal(t, 4, NewEditDistance("café", Damerau).Compare("", 10))
	require.Equal(t, 0, NewEditDistance("café", Damerau).Compare("café", 2))
	require.Equal(t, 1, NewEditDistance("café", Damerau).Compare("cafe", 2)) // é→e
	require.Equal(t, 1, NewEditDistance("abé", Damerau).Compare("aéb", 2))  // transpose

	// Non-BMP: 😀 is one code point → surrogate pair → Java length 2
	require.Equal(t, 2, NewEditDistance("", Damerau).Compare("😀", 10))
	require.Equal(t, 2, NewEditDistance("😀", Damerau).Compare("", 10))
	require.Equal(t, 0, NewEditDistance("😀", Damerau).Compare("😀", 2))
	// 😀a vs 😀b: last BMP unit differs → distance 1
	require.Equal(t, 1, NewEditDistance("😀a", Damerau).Compare("😀b", 2))
	// 😀=\uD83D\uDE00, 😁=\uD83D\uDE01: shared high surrogate → distance 1 (UTF-16 units)
	require.Equal(t, 1, NewEditDistance("😀", Damerau).Compare("😁", 2))
	require.Equal(t, 1, NewEditDistance("😀", Damerau).Compare("😁", 1))
	// Fully disjoint BMP runes: distance = unit length of longer after no shared affix
	require.Equal(t, 1, NewEditDistance("é", Damerau).Compare("e", 1))
	// Surrogate-free multi-unit substitute budget exceed
	require.Equal(t, -1, NewEditDistance("ab", Damerau).Compare("xy", 1))
	require.Equal(t, 2, NewEditDistance("ab", Damerau).Compare("xy", 2))
}

func TestEditDistance_SharedPrefixSuffix(t *testing.T) {
	// common prefix+suffix; single middle insert/substitute
	require.Equal(t, 1, NewEditDistance("prefix_suffix", Damerau).Compare("prefixXsuffix", 2))
	// all of shorter is prefix of longer → distance = length diff
	require.Equal(t, 2, NewEditDistance("ab", Damerau).Compare("abcd", 5))
	require.Equal(t, 2, NewEditDistance("abcd", Damerau).Compare("ab", 5))
	// identical after strip
	require.Equal(t, 0, NewEditDistance("aaa", Damerau).Compare("aaa", 5))
	// only suffix shared (first chars differ): "xbc" vs "abc"
	require.Equal(t, 1, NewEditDistance("xbc", Damerau).Compare("abc", 2))
	// only prefix shared
	require.Equal(t, 1, NewEditDistance("abx", Damerau).Compare("aby", 2))
	// full shorter is suffix of longer
	require.Equal(t, 2, NewEditDistance("ab", Damerau).Compare("xxab", 5))
	// ca → abc
	require.Equal(t, 3, NewEditDistance("ca", Damerau).Compare("abc", 3))
}

func TestEditDistance_MaxDistanceClampAndNegative(t *testing.T) {
	// maxDistance < 0 → clamp to tLen (Java)
	require.Equal(t, 0, NewEditDistance("same", Damerau).Compare("same", -1))
	require.Equal(t, 1, NewEditDistance("ab", Damerau).Compare("ac", -1))
	// maxDistance larger than tLen clamps; still returns true distance
	require.Equal(t, 1, NewEditDistance("hello", Damerau).Compare("hallo", 100))
}

func TestEditDistance_ReuseInstanceMultipleCompares(t *testing.T) {
	// v0/v2 resized and cleared across compares (Java instance fields)
	e := NewEditDistance("hi", Damerau)
	require.Equal(t, 0, e.Compare("hi", 2))
	require.Equal(t, 1, e.Compare("ho", 2))
	require.Equal(t, 2, e.Compare("xy", 2))
	require.Equal(t, -1, e.Compare("abcd", 1))
	// longer string2 forces v0 grow beyond ctor base length (hi → hixxxx = 4 inserts)
	require.Equal(t, 4, e.Compare("hixxxx", 10))
}

func TestEditDistance_UnknownAlgorithmPanics(t *testing.T) {
	// Java: throw new IllegalArgumentException("unknown DistanceAlgorithm")
	e := NewEditDistance("x", DistanceAlgorithm(99))
	require.PanicsWithValue(t, "unknown DistanceAlgorithm", func() {
		e.Compare("y", 1)
	})
}

func TestEditDistance_DamerauDirect(t *testing.T) {
	// Public method twin of Java DamerauLevenshteinDistance
	e := NewEditDistance("hello", Damerau)
	require.Equal(t, 0, e.DamerauLevenshteinDistance("hello", 2))
	require.Equal(t, 1, e.DamerauLevenshteinDistance("hallo", 2))
	require.Equal(t, -1, e.DamerauLevenshteinDistance("zzzzz", 1))
}

func TestEditDistance_CtorEmptyNoVTables(t *testing.T) {
	// empty base returns before allocating v0/v2 (Java early return)
	e := NewEditDistance("", Damerau)
	require.True(t, e.baseNull)
	require.Nil(t, e.v0)
	require.Nil(t, e.v2)
	require.Equal(t, 3, e.Compare("abc", 5))
}
