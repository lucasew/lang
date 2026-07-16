package hunspell

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHunspellNoSuggestionRule(t *testing.T) {
	d := NewMapHunspellDictionary([]string{"ok"})
	r := NewHunspellNoSuggestionRule(d)
	require.Equal(t, HunspellNoSuggestionRuleID, r.GetID())
	require.False(t, r.IsMisspelled("ok"))
	require.True(t, r.IsMisspelled("bad"))
	require.Nil(t, r.Suggest("bad"))
}
