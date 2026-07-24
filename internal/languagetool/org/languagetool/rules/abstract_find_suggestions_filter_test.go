package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func ptrFS(s string) *string { return &s }

func TestAbstractFindSuggestionsFilter_Diacritics(t *testing.T) {
	f := &AbstractFindSuggestionsFilter{
		SpellingSuggestions: func(atr *languagetool.AnalyzedTokenReadings) []string {
			return []string{"café", "cafe", "boat"}
		},
		Tag: func(word string) *languagetool.AnalyzedTokenReadings {
			// original "cafe" must NOT already match desiredPostag (else diacritics drops match)
			if word == "cafe" {
				return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(word, ptrFS("UNKNOWN"), nil))
			}
			if word == "boat" {
				return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(word, ptrFS("V"), nil))
			}
			return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(word, ptrFS("NCMS000"), nil))
		},
	}
	tok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("cafe", ptrFS("UNKNOWN"), nil))
	tok.SetStartPos(0)
	m := NewRuleMatch(NewFakeRule("R"), nil, 0, 4, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*", "Mode": "diacritics",
	}, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"café"}, out.GetSuggestedReplacements())
}

func TestAbstractFindSuggestionsFilter_Template(t *testing.T) {
	f := &AbstractFindSuggestionsFilter{
		SpellingSuggestions: func(atr *languagetool.AnalyzedTokenReadings) []string {
			return []string{"casa"}
		},
		Tag: func(word string) *languagetool.AnalyzedTokenReadings {
			return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(word, ptrFS("NCFS000"), nil))
		},
	}
	tok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("csa", ptrFS("NCFS000"), nil))
	tok.SetStartPos(0)
	m := NewRuleMatch(NewFakeRule("R"), nil, 0, 3, "msg")
	m.SetSuggestedReplacements([]string{"la {suggestion}"})
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*",
	}, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"la casa"}, out.GetSuggestedReplacements())
}

func TestAbstractFindSuggestionsFilter_SuppressEmpty(t *testing.T) {
	f := &AbstractFindSuggestionsFilter{
		SpellingSuggestions: func(atr *languagetool.AnalyzedTokenReadings) []string {
			return nil
		},
	}
	tok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("x", nil, nil))
	m := NewRuleMatch(NewFakeRule("R"), nil, 0, 1, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*", "suppressMatch": "true",
	}, []*languagetool.AnalyzedTokenReadings{tok}, nil))
}

func TestEqualWithoutDiacritics(t *testing.T) {
	// Java StringTools.removeDiacritics + equalsIgnoreCase (NFD Mn strip).
	require.True(t, equalWithoutDiacritics("café", "cafe"))
	require.True(t, equalWithoutDiacritics("CAFÉ", "cafe"))
	require.True(t, equalWithoutDiacritics("niño", "nino")) // ñ → n via NFD
	require.False(t, equalWithoutDiacritics("casa", "cosa"))
}

func TestStringComparatorLess_DamerauTransposition(t *testing.T) {
	// Damerau: "ab" → "ba" is distance 1 (transposition); plain Levenshtein invent is 2.
	// Closer candidate should sort first when only one is a transposition of word.
	require.True(t, stringComparatorLess("ab", "ba", "xy"),
		"transposition ba must be closer than unrelated xy")
	// Equal distances: comparator is strict less; equal → false either order is OK for sort stability
	require.False(t, stringComparatorLess("test", "test", "test"))
}

func TestAbstractFindSuggestionsFilter_RemoveSuggestionsRegexpCaseSensitive(t *testing.T) {
	// Java UNICODE_CASE without CASE_INSENSITIVE → case-sensitive matches().
	f := &AbstractFindSuggestionsFilter{
		SpellingSuggestions: func(atr *languagetool.AnalyzedTokenReadings) []string {
			return []string{"Bad", "good"}
		},
		Tag: func(word string) *languagetool.AnalyzedTokenReadings {
			return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(word, ptrFS("NCMS000"), nil))
		},
	}
	tok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("gxd", ptrFS("UNKNOWN"), nil))
	tok.SetStartPos(0)
	m := NewRuleMatch(NewFakeRule("R"), nil, 0, 3, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*", "removeSuggestionsRegexp": "bad",
	}, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	// "Bad" must NOT be removed by case-insensitive invent of "bad"
	require.Equal(t, []string{"Bad", "good"}, out.GetSuggestedReplacements())
}

func atrFS(surface string, start int) *languagetool.AnalyzedTokenReadings {
	atr := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(surface, nil, nil))
	atr.SetStartPos(start)
	return atr
}

// Twin: RuleFilter.getPosition — marker, marker±N, numeric, OOB panic (no skipCorrected invent).
func TestGetPositionFS_MarkerAndOffset(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atrFS("the", 0),
		atrFS("runing", 4),
		atrFS("fast", 11),
	}
	m := NewRuleMatch(NewFakeRule("R"), nil, 4, 10, "msg")
	// marker → first token at/after FromPos → index 1 ("runing")
	require.Equal(t, 1, getPositionFS("marker", toks, m))
	// marker+1 → one past marker token → "fast"
	require.Equal(t, 2, getPositionFS("marker+1", toks, m))
	// marker-1 → token before marker → "the"
	require.Equal(t, 0, getPositionFS("marker-1", toks, m))
	// numeric 1-based
	require.Equal(t, 0, getPositionFS("1", toks, m))
	require.Equal(t, 1, getPositionFS("2", toks, m))
	// OOB → panic like IllegalArgumentException
	require.Panics(t, func() { getPositionFS("0", toks, m) })
	require.Panics(t, func() { getPositionFS("4", toks, m) })
	require.Panics(t, func() { getPositionFS("marker+9", toks, m) })
}

// Twin: definitiveReplacements.remove(match.getOriginalErrorStr()) only —
// do not invent setOriginalErrorStr when empty.
// Pre-seed match suggested replacements (non-template path) so remove is observable;
// spelling path already skips suggestion.equals(token).
func TestAbstractFindSuggestionsFilter_OriginalErrorStrRemoveOnly(t *testing.T) {
	f := &AbstractFindSuggestionsFilter{
		SpellingSuggestions: func(atr *languagetool.AnalyzedTokenReadings) []string {
			return []string{"running"}
		},
		Tag: func(word string) *languagetool.AnalyzedTokenReadings {
			return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(word, ptrFS("VBG"), nil))
		},
	}
	tok := atrFS("runing", 0)
	// Hand-built match: OriginalErrorStr unset (""). Java remove("") does not strip "runing".
	// Inventing SetOriginalErrorStr() from sentence span would wrongly drop it.
	m := NewRuleMatch(NewFakeRule("R"), nil, 0, 6, "msg")
	m.SetSuggestedReplacements([]string{"runing", "also"})
	require.Equal(t, "", m.GetOriginalErrorStr())
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "VB.*",
	}, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetSuggestedReplacements(), "runing")
	require.Contains(t, out.GetSuggestedReplacements(), "also")
	require.Contains(t, out.GetSuggestedReplacements(), "running")
	require.Equal(t, "", m.GetOriginalErrorStr(), "must not invent SetOriginalErrorStr")

	// When OriginalErrorStr is set (pattern matcher path), remove that surface only.
	m2 := NewRuleMatch(NewFakeRule("R"), nil, 0, 6, "msg")
	m2.OriginalErrorStr = "runing"
	m2.SetSuggestedReplacements([]string{"runing", "also"})
	out2 := f.AcceptRuleMatch(m2, map[string]string{
		"wordFrom": "1", "desiredPostag": "VB.*",
	}, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out2)
	require.NotContains(t, out2.GetSuggestedReplacements(), "runing")
	require.Contains(t, out2.GetSuggestedReplacements(), "also")
	require.Contains(t, out2.GetSuggestedReplacements(), "running")
}
