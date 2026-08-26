package multitoken

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/broker"
	"github.com/stretchr/testify/require"
)

// Twin: exact dictionary hit → stopSearching → empty suggestions.
func TestMultitokenSpeller_ExactMatchStops(t *testing.T) {
	m := NewMultitokenSpeller()
	require.NoError(t, m.LoadWords(strings.NewReader(`# comment
New York
Los Angeles
`)))
	require.Empty(t, m.GetSuggestions("New York"))
	// lower query hits noSpaces key with weight 0
	sugg := m.GetSuggestions("new york")
	require.Contains(t, sugg, "New York")
}

// Twin: single-token lines ignored (Java continue when !normalizedKey.contains(" ")).
func TestMultitokenSpeller_SkipsSingleTokenLines(t *testing.T) {
	m := NewMultitokenSpeller()
	require.NoError(t, m.LoadWords(strings.NewReader("hello\nhello world\n")))
	require.Contains(t, m.GetSuggestions("helo world"), "hello world")
	// single-token "hello" never entered the map
	require.Empty(t, m.GetSuggestions("helo"))
}

// Twin: # comments and trailing comment strip (substringBefore + trim).
func TestMultitokenSpeller_CommentHandling(t *testing.T) {
	m := NewMultitokenSpeller()
	require.NoError(t, m.LoadWords(strings.NewReader(`# full line comment
foo bar # trailing
`)))
	require.Contains(t, m.GetSuggestions("foo baz"), "foo bar")
}

// Twin: typo suggestion via first-char map + distancesPerWord.
func TestMultitokenSpeller_Typo(t *testing.T) {
	m := NewMultitokenSpeller()
	require.NoError(t, m.LoadWords(strings.NewReader("hello world\n")))
	sugg := m.GetSuggestions("helo world")
	require.Equal(t, []string{"hello world"}, sugg)
}

// Twin: MultitokenSpeller.getNormalizeKey
func TestGetNormalizeKey(t *testing.T) {
	require.Equal(t, "cafe latte", getNormalizeKey("Café-Latte"))
	// Java does not collapse internal double spaces after hyphen replace
	require.Equal(t, "a  b", getNormalizeKey("a--b"))
}

// Twin: MultitokenSpeller first-char map key is String.charAt(0) (UTF-16 unit).
func TestFirstUTF16UnitAsRune(t *testing.T) {
	require.Equal(t, 'n', firstUTF16UnitAsRune("new york"))
	require.Equal(t, 'N', firstUTF16UnitAsRune("New York"))
	require.Equal(t, 'c', firstUTF16UnitAsRune("café latte"))
	require.Equal(t, rune(0), firstUTF16UnitAsRune(""))
}

// Twin of MultitokenSpeller.WHITESPACE_AND_SEP = \p{Zs}+
func TestCollapseWhitespace_ZsOnly(t *testing.T) {
	require.Equal(t, "a b", collapseWhitespace("a\u00a0\u00a0b"))
	require.Equal(t, "a b", collapseWhitespace("a  b"))
	// leading/trailing Zs become single spaces (Java replaceAll, not trim)
	require.Equal(t, " a b ", collapseWhitespace("  a  b  "))
	// tab is NOT Zs — must remain
	require.Equal(t, "a\tb", collapseWhitespace("a\tb"))
	require.Equal(t, "a\nb", collapseWhitespace("a\nb"))
}

// Twin: per-token distance path (CA-style "Manuel Sadusky" → "Manuel Sadosky").
func TestMultitokenSpeller_PerTokenDistance(t *testing.T) {
	m := NewMultitokenSpeller()
	require.NoError(t, m.LoadWords(strings.NewReader("Manuel Sadosky\n")))
	require.Equal(t, []string{"Manuel Sadosky"}, m.GetSuggestions("Manuel Sadusky"))
	require.Equal(t, []string{"Manuel Sadosky"}, m.GetSuggestions("Manue Sadosky"))
}

// Twin: discardRunOnWords with SpellingCheckRule.isMisspelled hook.
func TestMultitokenSpeller_DiscardRunOnWords(t *testing.T) {
	m := NewMultitokenSpeller()
	m.IsMisspelledToken = func(tok string) bool {
		return tok != "hell" && tok != "oworld" && tok != "hello" && tok != "world"
	}
	// sugg2a=hello, sugg2b=world — both known → discard true
	require.True(t, m.discardRunOnWords("hell oworld"))
	// Capitalized second token never discarded
	require.False(t, m.discardRunOnWords("hell Oworld"))
	// nil speller hook → false (cannot discard)
	m2 := NewMultitokenSpeller()
	require.False(t, m2.discardRunOnWords("hell oworld"))
}

// Twin: discardRunOnWords causes computeSuggestions to return empty.
func TestMultitokenSpeller_DiscardRunOnEmptiesSuggestions(t *testing.T) {
	m := NewMultitokenSpeller()
	require.NoError(t, m.LoadWords(strings.NewReader("hello world\n")))
	m.IsMisspelledToken = func(tok string) bool {
		return tok != "hell" && tok != "oworld" && tok != "hello" && tok != "world"
	}
	require.Empty(t, m.GetSuggestions("hell oworld"))
}

// Twin: levenshteinDistance space-strip, anagram, normalizeSimilarChars.
func TestLevenshteinDistance_AnagramAndSimilar(t *testing.T) {
	require.Equal(t, 0, levenshteinDistance("ab c", "a bc"))
	require.Equal(t, 1, levenshteinDistance("ab", "ba"))
	// y→i similar chars
	require.Equal(t, 0, levenshteinDistance("by", "bi"))
}

// Twin: maxEditDistance with correct-token reduction.
func TestMaxEditDistance_CorrectTokens(t *testing.T) {
	// "hello world" vs "hello wrld" — correct "hello" = 5 chars
	// totalLength 11, correct 5 → correctLength 6 → <=7 → 2 - firstCharWrong
	require.Equal(t, 2, maxEditDistance("hello world", "hello wrld"))
}

// Twin: firstCharacterDistances charDistance pairs.
func TestCharDistance(t *testing.T) {
	require.Equal(t, 0.0, charDistance('a', 'a'))
	require.Equal(t, 0.2, charDistance('s', 'z'))
	require.Equal(t, 0.2, charDistance('v', 'b'))
	require.Equal(t, 0.0, charDistance('i', 'y'))
	require.Equal(t, 1.0, charDistance('a', 'x'))
}

// Twin: areTokensAcceptedBySpeller gates (all-upper first candidate / weight>1).
// Note: Java cache key is only the whitespace-normalized word (not the accepted flag),
// so a second call with a different accepted flag reuses the first computation.
func TestMultitokenSpeller_AreTokensAcceptedBySpeller(t *testing.T) {
	m := NewMultitokenSpeller()
	require.NoError(t, m.LoadWords(strings.NewReader("hello world\n")))
	require.Equal(t, []string{"hello world"}, m.GetSuggestionsOpts("helo world", false))
	// with accepted, weight 1 is allowed (<=1) — fresh speller to avoid cache
	m2 := NewMultitokenSpeller()
	require.NoError(t, m2.LoadWords(strings.NewReader("hello world\n")))
	require.Equal(t, []string{"hello world"}, m2.GetSuggestionsOpts("helo world", true))

	// all-upper original matching upper of first candidate → empty when accepted
	m3 := NewMultitokenSpeller()
	require.NoError(t, m3.LoadWords(strings.NewReader("hello world\n")))
	require.Empty(t, m3.GetSuggestionsOpts("HELLO WORLD", true))

	// without accepted flag, all-upper still gets suggestion (fresh cache)
	m4 := NewMultitokenSpeller()
	require.NoError(t, m4.LoadWords(strings.NewReader("hello world\n")))
	require.Equal(t, []string{"hello world"}, m4.GetSuggestionsOpts("HELLO WORLD", false))
}

// Twin: getAdditionalSuggestions merge + exact-hit empty return.
func TestMultitokenSpeller_AdditionalSuggestions(t *testing.T) {
	m := NewMultitokenSpeller()
	m.GetAdditionalSuggestions = func(w string) []WeightedSuggestion {
		return []WeightedSuggestion{NewWeightedSuggestion("foo bar", 0)}
	}
	require.Equal(t, []string{"foo bar"}, m.GetSuggestions("foo baz"))

	m.GetAdditionalSuggestions = func(w string) []WeightedSuggestion {
		return []WeightedSuggestion{NewWeightedSuggestion(w, 0)}
	}
	require.Empty(t, m.GetSuggestions("exact match"))
}

// Twin: isException hook stops search.
func TestMultitokenSpeller_IsException(t *testing.T) {
	m := NewMultitokenSpeller()
	require.NoError(t, m.LoadWords(strings.NewReader("hello world\n")))
	m.IsException = func(original, candidate string) bool {
		return original == "helo world" && candidate == "hello world"
	}
	require.Empty(t, m.GetSuggestions("helo world"))
}

// Twin: hyphen spacing normalize before lookup ("- " / " -" → "-").
// stopSearching compares originalWord (pre-hyphen-fix) to candidates, so a
// spaced-hyphen query is not an exact stop even if the hyphen-fixed form matches.
func TestMultitokenSpeller_HyphenSpacing(t *testing.T) {
	m := NewMultitokenSpeller()
	require.NoError(t, m.LoadWords(strings.NewReader("well-known fact\n")))
	// exact original == candidate → stopSearching empty
	require.Empty(t, m.GetSuggestions("well-known fact"))
	// "well -known fact" → word "well-known fact" → normalize key matches noSpaces;
	// original ≠ candidate → suggestions returned (Java same)
	require.Equal(t, []string{"well-known fact"}, m.GetSuggestions("well -known fact"))
}

// Twin: stopSearching title-case of all-lower candidate.
func TestMultitokenSpeller_TitleCaseStop(t *testing.T) {
	m := NewMultitokenSpeller()
	require.NoError(t, m.LoadWords(strings.NewReader("hello world\n")))
	// candidate is all-lower; title case "Hello World" == original → stop
	require.Empty(t, m.GetSuggestions("Hello World"))
}

// Twin: cache key is whitespace-normalized word only (same result for same key).
func TestMultitokenSpeller_Cache(t *testing.T) {
	m := NewMultitokenSpeller()
	require.NoError(t, m.LoadWords(strings.NewReader("hello world\n")))
	a := m.GetSuggestions("helo world")
	b := m.GetSuggestions("helo world")
	require.Equal(t, a, b)
	require.Equal(t, []string{"hello world"}, a)
}

// Twin: PrepareLine hook (language.prepareLineForSpeller).
func TestMultitokenSpeller_PrepareLine(t *testing.T) {
	m := NewMultitokenSpeller()
	m.PrepareLine = func(line string) []string {
		if strings.HasSuffix(line, "\tVB") {
			return []string{""}
		}
		if i := strings.Index(line, "\t"); i >= 0 {
			return []string{line[:i]}
		}
		return []string{line}
	}
	require.NoError(t, m.LoadWords(strings.NewReader("New York\tNNP\nfoo bar\tVB\n")))
	require.Equal(t, []string{"New York"}, m.GetSuggestions("New Yrok"))
	require.Empty(t, m.GetSuggestions("foo bar"))
}

// Twin: numberOfCorrectChars uses Java String.split (keeps internal empties).
func TestNumberOfCorrectChars(t *testing.T) {
	require.Equal(t, 5, numberOfCorrectChars("hello world", "hello wrld")) // "hello"=5
	// parts equal length 2: hello!=hallo, world==world → 5
	require.Equal(t, 5, numberOfCorrectChars("hello world", "hallo world"))
	// single token → 0
	require.Equal(t, 0, numberOfCorrectChars("hello", "hello"))
}

// Twin: short candidates (<7 UTF-16 units) only allow distance 0.
func TestMultitokenSpeller_ShortCandidateRequiresDistance0(t *testing.T) {
	m := NewMultitokenSpeller()
	// "ab cd" normalized length 5 < 7
	require.NoError(t, m.LoadWords(strings.NewReader("ab cd\n")))
	// distance > 0 typo should not suggest via first-char map
	require.Empty(t, m.GetSuggestions("ax cd"))
	// casing via noSpaces: "AB CD" → normalize "ab cd" noSpaces hit weight 0
	require.Equal(t, []string{"ab cd"}, m.GetSuggestions("AB CD"))
}

// Twin: NewMultitokenSpellerFromResources loads broker paths + once-guard.
func TestNewMultitokenSpellerFromResources(t *testing.T) {
	b := broker.NewMapResourceDataBroker()
	b.Resource["xx/multiwords.txt"] = "alpha beta\n"
	sp, err := NewMultitokenSpellerFromResources(nil, nil, b, []string{"/xx/multiwords.txt"})
	require.NoError(t, err)
	require.Equal(t, []string{"alpha beta"}, sp.GetSuggestions("alpha betx"))

	// once-guard: second init does not reload / wipe
	b.Resource["xx/other.txt"] = "gamma delta\n"
	require.NoError(t, sp.InitMultitokenSpeller(b, []string{"/xx/other.txt"}))
	require.Equal(t, []string{"alpha beta"}, sp.GetSuggestions("alpha betx"))
	// other.txt not loaded due to once-guard
	require.Empty(t, sp.GetSuggestions("gamma deltx"))
}

// Twin: splitBySpace omits empties (StringUtils.split); javaSplitOnSpace keeps internal.
func TestSplitBySpace_VsJavaSplit(t *testing.T) {
	require.Equal(t, []string{"a", "b"}, splitBySpace("a  b"))
	require.Equal(t, []string{"a", "", "b"}, javaSplitOnSpace("a  b"))
	require.Equal(t, []string{"a", "b"}, javaSplitOnSpace("a b "))
}

// Twin: distancesPerWord equal multi-token vs unequal.
func TestDistancesPerWord(t *testing.T) {
	// equal part counts >1 → per-token
	d := distancesPerWord([]string{"hello", "world"}, []string{"helo", "world"}, "hello world", "helo world")
	require.Len(t, d, 2)
	require.Equal(t, 1, d[0])
	require.Equal(t, 0, d[1])
	// unequal → whole-string
	d2 := distancesPerWord([]string{"a", "b"}, []string{"abc"}, "a b", "abc")
	require.Len(t, d2, 1)
}

// Twin: normalizeSimilarChars sequential replaces.
func TestNormalizeSimilarChars(t *testing.T) {
	require.Equal(t, "bi", normalizeSimilarChars("by"))
	require.Equal(t, "co", normalizeSimilarChars("ko"))
	require.Equal(t, "ca", normalizeSimilarChars("ka"))
	require.Equal(t, "cati", normalizeSimilarChars("katy"))
}
