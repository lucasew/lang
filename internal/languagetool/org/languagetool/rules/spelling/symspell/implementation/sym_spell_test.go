package implementation

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// fixtureDict builds the multi-word dictionary used for Java golden comparisons.
// Java create order/returns: hello 10→true, world 5→true, hello 1→false (count=11),
// hallo 3→true, help 8→true, café 10→true; maxLength=5.
func fixtureDict(t *testing.T) *SymSpell {
	t.Helper()
	s := NewSymSpell(16, 2, 7, 1)
	require.True(t, s.CreateDictionaryEntry("hello", 10, nil))
	require.True(t, s.CreateDictionaryEntry("world", 5, nil))
	require.False(t, s.CreateDictionaryEntry("hello", 1, nil)) // update only
	require.True(t, s.CreateDictionaryEntry("hallo", 3, nil))
	require.True(t, s.CreateDictionaryEntry("help", 8, nil))
	require.True(t, s.CreateDictionaryEntry("café", 10, nil))
	require.Equal(t, 5, s.maxLength)
	require.Equal(t, 5, s.WordCount())
	return s
}

// Twin: constructor clamps negative / invalid args to Java defaults;
// compactMask uses signed 0xffffffff arithmetic → -4.
func TestNewSymSpell_Defaults(t *testing.T) {
	s := NewSymSpell(-1, -1, 0, -1)
	require.Equal(t, 2, s.maxDictionaryEditDistance)
	require.Equal(t, 7, s.prefixLength)
	require.Equal(t, int64(1), s.countThreshold)
	require.Equal(t, 16, s.initialCapacity)
	require.Equal(t, -4, s.compactMask)
	require.Nil(t, s.deletes)
	require.Equal(t, Damerau, s.distanceAlgorithm)
}

// Twin: getStringHash values from Java SymSpell with defaultCompactLevel=5.
func TestGetStringHash_JavaValues(t *testing.T) {
	s := DefaultSymSpell()
	require.Equal(t, -4, s.compactMask)
	// Captured from JVM getStringHash:
	require.Equal(t, -2128831036, s.getStringHash(""))
	require.Equal(t, -468965075, s.getStringHash("a"))
	require.Equal(t, 1294271946, s.getStringHash("ab"))
	require.Equal(t, 1335831723, s.getStringHash("hello"))
	require.Equal(t, 933488787, s.getStringHash("world"))
	require.Equal(t, 856211071, s.getStringHash("café"))
	require.Equal(t, 460420691, s.getStringHash("helo"))
}

func TestSymSpellExactAndTypo(t *testing.T) {
	s := fixtureDict(t)

	// exact
	got := s.Lookup("hello", VerbosityTop)
	require.Len(t, got, 1)
	require.Equal(t, "hello", got[0].Term)
	require.Equal(t, 0, got[0].Distance)
	require.Equal(t, int64(11), got[0].Count)

	// single edit — Java Top: hello|1|11
	got = s.Lookup("helo", VerbosityTop)
	require.Len(t, got, 1)
	require.Equal(t, "hello", got[0].Term)
	require.Equal(t, 1, got[0].Distance)
	require.Equal(t, int64(11), got[0].Count)
}

// Twin: Verbosity.Top / Closest / All filtering vs Java goldens on "helo" / "hell" / "hello".
func TestLookup_VerbosityFiltering(t *testing.T) {
	s := fixtureDict(t)

	// helo Top => [hello|1|11]
	top := s.Lookup("helo", VerbosityTop)
	require.Len(t, top, 1)
	require.Equal(t, SuggestItem{Term: "hello", Distance: 1, Count: 11}, top[0])

	// helo Closest => [hello|1|11, help|1|8]  (same ed, count desc)
	closest := s.Lookup("helo", VerbosityClosest)
	require.Equal(t, []SuggestItem{
		{Term: "hello", Distance: 1, Count: 11},
		{Term: "help", Distance: 1, Count: 8},
	}, closest)

	// helo All => [hello|1|11, help|1|8, hallo|2|3]
	all := s.Lookup("helo", VerbosityAll)
	require.Equal(t, []SuggestItem{
		{Term: "hello", Distance: 1, Count: 11},
		{Term: "help", Distance: 1, Count: 8},
		{Term: "hallo", Distance: 2, Count: 3},
	}, all)

	// hello All includes exact + neighbors: [hello|0|11, hallo|1|3, help|2|8]
	allExact := s.Lookup("hello", VerbosityAll)
	require.Equal(t, []SuggestItem{
		{Term: "hello", Distance: 0, Count: 11},
		{Term: "hallo", Distance: 1, Count: 3},
		{Term: "help", Distance: 2, Count: 8},
	}, allExact)

	// hell Top/Closest/All (Java goldens)
	require.Equal(t, []SuggestItem{{Term: "hello", Distance: 1, Count: 11}}, s.Lookup("hell", VerbosityTop))
	require.Equal(t, []SuggestItem{
		{Term: "hello", Distance: 1, Count: 11},
		{Term: "help", Distance: 1, Count: 8},
	}, s.Lookup("hell", VerbosityClosest))
	require.Equal(t, []SuggestItem{
		{Term: "hello", Distance: 1, Count: 11},
		{Term: "help", Distance: 1, Count: 8},
		{Term: "hallo", Distance: 2, Count: 3},
	}, s.Lookup("hell", VerbosityAll))
}

// Twin: no suggestions when input cannot match.
func TestLookup_NoMatch(t *testing.T) {
	s := fixtureDict(t)
	require.Empty(t, s.Lookup("xyz", VerbosityAll))
}

// Twin: early exit when inputLen - maxEditDistance > maxLength.
func TestLookup_InputTooLong(t *testing.T) {
	s := NewSymSpell(16, 2, 7, 1)
	require.True(t, s.CreateDictionaryEntry("ab", 1, nil))
	require.Empty(t, s.Lookup("abcdefghijklmnop", VerbosityAll))
}

// Twin: maxEditDistance > maxDictionaryEditDistance → IllegalArgumentException message.
func TestLookup_DistTooBig(t *testing.T) {
	s := DefaultSymSpell()
	require.PanicsWithValue(t, "Dist to big: 99", func() {
		s.LookupMax("x", VerbosityTop, 99)
	})
}

func TestSymSpellStaging(t *testing.T) {
	s := DefaultSymSpell()
	st := NewSuggestionStage(16)
	require.True(t, s.CreateDictionaryEntry("test", 3, st))
	require.True(t, s.CreateDictionaryEntry("toast", 2, st))
	// deletes still nil until commit (Java)
	require.Nil(t, s.deletes)
	s.CommitStaging(st)
	got := s.Lookup("test", VerbosityTop)
	require.Equal(t, []SuggestItem{{Term: "test", Distance: 0, Count: 3}}, got)
	got = s.Lookup("toas", VerbosityClosest)
	require.Equal(t, []SuggestItem{{Term: "toast", Distance: 1, Count: 2}}, got)
}

func TestSymSpellBelowThreshold(t *testing.T) {
	s := NewSymSpell(16, 2, 7, 3)
	require.False(t, s.CreateDictionaryEntry("rare", 1, nil))
	require.Equal(t, 0, s.WordCount())
	require.False(t, s.CreateDictionaryEntry("rare", 1, nil))
	require.True(t, s.CreateDictionaryEntry("rare", 1, nil)) // reaches 3
	require.Equal(t, 1, s.WordCount())
	got := s.Lookup("rare", VerbosityTop)
	require.Equal(t, []SuggestItem{{Term: "rare", Distance: 0, Count: 3}}, got)
}

// Java String.length for "café" is 4 (UTF-16), not 5 UTF-8 bytes.
func TestSymSpellAccentedUTF16(t *testing.T) {
	s := DefaultSymSpell()
	require.True(t, s.CreateDictionaryEntry("café", 10, nil))
	got := s.Lookup("café", VerbosityTop)
	require.Equal(t, []SuggestItem{{Term: "café", Distance: 0, Count: 10}}, got)
	// Java: cafe Closest => [café|1|10]
	got = s.Lookup("cafe", VerbosityClosest)
	require.Equal(t, []SuggestItem{{Term: "café", Distance: 1, Count: 10}}, got)
	require.Equal(t, 4, s.maxLength)
}

// Twin: count overflow saturates at Long.MAX_VALUE.
func TestCreateDictionaryEntry_CountSaturate(t *testing.T) {
	s := DefaultSymSpell()
	require.True(t, s.CreateDictionaryEntry("big", math.MaxInt64-5, nil))
	require.False(t, s.CreateDictionaryEntry("big", 100, nil)) // update
	require.Equal(t, int64(math.MaxInt64), s.words["big"])
}

// Twin: count <= 0 with threshold > 0 returns false without adding.
func TestCreateDictionaryEntry_NonPositiveCount(t *testing.T) {
	s := DefaultSymSpell()
	require.False(t, s.CreateDictionaryEntry("x", 0, nil))
	require.False(t, s.CreateDictionaryEntry("x", -3, nil))
	require.Equal(t, 0, s.WordCount())
}
