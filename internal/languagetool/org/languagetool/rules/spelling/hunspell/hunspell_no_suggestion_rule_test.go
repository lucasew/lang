package hunspell

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestHunspellNoSuggestionRule(t *testing.T) {
	d := NewMapHunspellDictionary([]string{"ok"})
	r := NewHunspellNoSuggestionRule("is", d)
	require.Equal(t, HunspellNoSuggestionRuleID, r.GetID())
	require.False(t, r.IsMisspelledWord("ok"))
	require.True(t, r.IsMisspelledWord("bad"))
	require.Nil(t, r.Suggest("bad"))

	// Match flags misspellings but never attaches suggestions.
	sent := languagetool.AnalyzePlain("bad ok")
	ms, err := r.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	for _, m := range ms {
		require.Empty(t, m.GetSuggestedReplacements())
	}
}
