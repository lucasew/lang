package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAbstractSuppressMisspelledSuggestionsFilter(t *testing.T) {
	f := &AbstractSuppressMisspelledSuggestionsFilter{
		IsMisspelled: func(s string) bool { return s == "teh" },
	}
	m := NewRuleMatch(NewFakeRule("R"), nil, 0, 3, "msg")
	m.SetSuggestedReplacements([]string{"the", "teh", "that"})
	out := f.AcceptRuleMatch(m, map[string]string{"suppressMatch": "true"})
	require.NotNil(t, out)
	require.Equal(t, []string{"the", "that"}, out.GetSuggestedReplacements())

	m2 := NewRuleMatch(NewFakeRule("R"), nil, 0, 3, "msg")
	m2.SetSuggestedReplacements([]string{"teh"})
	require.Nil(t, f.AcceptRuleMatch(m2, map[string]string{"suppressMatch": "true"}))
	// keep match with empty suggestions if suppressMatch=false
	m3 := NewRuleMatch(NewFakeRule("R"), nil, 0, 3, "msg")
	m3.SetSuggestedReplacements([]string{"teh"})
	out3 := f.AcceptRuleMatch(m3, map[string]string{"suppressMatch": "false"})
	require.NotNil(t, out3)
	require.Empty(t, out3.GetSuggestedReplacements())
}

// Twin of AbstractSuppressMisspelledSuggestionsFilter.isMisspelled:
// WordTokenizer splits suggestion; each token is checked with the speller.
func TestAbstractSuppressMisspelled_TokenizesSuggestion(t *testing.T) {
	// Speller only knows whole-phrase invent would fail; per-token "house" is OK.
	// "house!" → WordTokenizer → ["house", "!"] — neither is "teh".
	f := &AbstractSuppressMisspelledSuggestionsFilter{
		IsMisspelled: func(s string) bool { return s == "teh" || s == "xyz" },
	}
	m := NewRuleMatch(NewFakeRule("R"), nil, 0, 3, "msg")
	m.SetSuggestedReplacements([]string{"house!", "teh house", "xyz"})
	out := f.AcceptRuleMatch(m, map[string]string{"suppressMatch": "true"})
	require.NotNil(t, out)
	// "house!" tokenizes to house + ! → keep
	// "teh house" has token "teh" → drop
	// "xyz" → drop
	require.Equal(t, []string{"house!"}, out.GetSuggestedReplacements())
}

// Injected Tokenize mirrors language-specific WordTokenizer.
func TestAbstractSuppressMisspelled_CustomTokenize(t *testing.T) {
	f := &AbstractSuppressMisspelledSuggestionsFilter{
		IsMisspelled: func(s string) bool { return s == "bad" },
		Tokenize: func(s string) []string {
			if s == "good-bad" {
				return []string{"good", "bad"}
			}
			return []string{s}
		},
	}
	m := NewRuleMatch(NewFakeRule("R"), nil, 0, 3, "msg")
	m.SetSuggestedReplacements([]string{"good-bad", "good"})
	out := f.AcceptRuleMatch(m, map[string]string{"suppressMatch": "true"})
	require.NotNil(t, out)
	require.Equal(t, []string{"good"}, out.GetSuggestedReplacements())
}

// Twin of Java RuleFilter.getRequired("suppressMatch") — throws when absent.
func TestAbstractSuppressMisspelled_RequiresSuppressMatch(t *testing.T) {
	f := &AbstractSuppressMisspelledSuggestionsFilter{}
	m := NewRuleMatch(NewFakeRule("R"), nil, 0, 3, "msg")
	m.SetSuggestedReplacements([]string{"ok"})
	require.PanicsWithValue(t, "Missing key 'suppressMatch'", func() {
		f.AcceptRuleMatch(m, map[string]string{})
	})
	require.PanicsWithValue(t, "Missing key 'suppressMatch'", func() {
		f.AcceptRuleMatch(m, nil)
	})
}
