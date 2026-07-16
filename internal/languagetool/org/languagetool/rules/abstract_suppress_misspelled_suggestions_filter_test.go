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
