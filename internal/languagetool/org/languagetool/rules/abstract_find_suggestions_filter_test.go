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
