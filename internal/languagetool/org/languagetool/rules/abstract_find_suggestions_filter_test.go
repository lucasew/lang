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
	require.True(t, equalWithoutDiacritics("café", "cafe"))
	require.False(t, equalWithoutDiacritics("casa", "cosa"))
}
