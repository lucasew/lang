package hunspell

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestHunspellRuleMatch(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"hello", "world"})
	dict.SetSuggestions("helo", []string{"hello"})
	r := NewHunspellRule("en", dict)
	require.Equal(t, HunspellRuleID, r.GetID())
	require.True(t, r.IsMisspelledWord("helo"))
	require.False(t, r.IsMisspelledWord("hello"))

	sent := languagetool.AnalyzePlain("hello helo world")
	matches, err := r.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	require.Equal(t, []string{"hello"}, matches[0].GetSuggestedReplacements())
}
