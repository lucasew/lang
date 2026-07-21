package multitoken

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultitokenSpeller_Exact(t *testing.T) {
	m := NewMultitokenSpeller()
	require.NoError(t, m.LoadWords(strings.NewReader(`# c
New York
Los Angeles
`)))
	// exact match → stop (no suggestions)
	require.Empty(t, m.GetSuggestions("New York"))
	// diacritic/case normalize exact via no-spaces? "new york" lower equals dict after normalize
	// original "new york" != "New York" but stopSearching title case of lower candidate:
	// candidate "New York" is not all-lower; second loop checks candidate==toLower → false
	// so we get suggestions from noSpaces key
	sugg := m.GetSuggestions("new york")
	require.Contains(t, sugg, "New York")
}

func TestMultitokenSpeller_Typo(t *testing.T) {
	m := NewMultitokenSpeller()
	require.NoError(t, m.LoadWords(strings.NewReader("hello world\n")))
	sugg := m.GetSuggestions("helo world")
	require.Contains(t, sugg, "hello world")
}

func TestGetNormalizeKey(t *testing.T) {
	require.Equal(t, "cafe latte", getNormalizeKey("Café-Latte"))
}

// Twin of MultitokenSpeller.WHITESPACE_AND_SEP = \p{Zs}+
func TestCollapseWhitespace_ZsOnly(t *testing.T) {
	// NBSP (\u00a0) is Zs → single space
	require.Equal(t, "a b", collapseWhitespace("a\u00a0\u00a0b"))
	// multiple ASCII spaces
	require.Equal(t, "a b", collapseWhitespace("a  b"))
	// leading/trailing Zs become single spaces (Java replaceAll, not trim)
	require.Equal(t, " a b ", collapseWhitespace("  a  b  "))
	// tab is NOT Zs — must remain (strings.Fields would remove it)
	require.Equal(t, "a\tb", collapseWhitespace("a\tb"))
	// newline is not Zs
	require.Equal(t, "a\nb", collapseWhitespace("a\nb"))
}

// Twin of MultitokenSpeller.distancesPerWord + per-token max distance.
func TestMultitokenSpeller_PerTokenDistance(t *testing.T) {
	m := NewMultitokenSpeller()
	require.NoError(t, m.LoadWords(strings.NewReader("Manuel Sadosky\n")))
	// Java CA test: "Manuel Sadusky" → [Manuel Sadosky]
	sugg := m.GetSuggestions("Manuel Sadusky")
	require.Contains(t, sugg, "Manuel Sadosky")

	// First-token typo only
	sugg2 := m.GetSuggestions("Manue Sadosky")
	require.Contains(t, sugg2, "Manuel Sadosky")
}

func TestMultitokenSpeller_DiscardRunOnWords(t *testing.T) {
	m := NewMultitokenSpeller()
	// Known words: "hell" and "oworld" invented as correctly spelled
	m.IsMisspelledToken = func(tok string) bool {
		return tok != "hell" && tok != "oworld" && tok != "hello" && tok != "world"
	}
	// "hello world" with space mis-split? "hell oworld" → discard if both parts known
	// parts[0]=hell, parts[1]=oworld → sugg1a=hel, sugg1b=loworld — misspelled
	// sugg2a=hello, sugg2b=world — both known → discard true
	require.True(t, m.discardRunOnWords("hell oworld"))

	// Capitalized second token never discarded
	require.False(t, m.discardRunOnWords("hell Oworld"))
}

func TestLevenshteinDistance_AnagramAndSimilar(t *testing.T) {
	// same after space strip
	require.Equal(t, 0, levenshteinDistance("ab c", "a bc"))
	// anagram same length → distance 1 when > 0
	d := levenshteinDistance("ab", "ba")
	require.Equal(t, 1, d)
}

func TestMaxEditDistance_CorrectTokens(t *testing.T) {
	// one correct long token reduces correctLength
	// "hello world" vs "hello wrld" — correct "hello" = 5 chars
	// totalLength 11, correct 5 → correctLength 6 → still <=7 → 2 - firstCharWrong
	// first chars h/h and w/w → 0
	require.Equal(t, 2, maxEditDistance("hello world", "hello wrld"))
}

func TestGetNormalizeKey_NoCollapse(t *testing.T) {
	// Java does not collapse internal double spaces after hyphen replace
	require.Equal(t, "cafe latte", getNormalizeKey("Café-Latte"))
}
