package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestEnglishAdditionalTopSuggestions_CaseSensitive(t *testing.T) {
	got := EnglishAdditionalTopSuggestions("Ths", nil)
	require.Equal(t, []string{"This", "The"}, got)
	// lowercase key only in ignoreCase map
	require.Equal(t, []string{"Jason"}, EnglishAdditionalTopSuggestions("json", nil))
	require.Equal(t, []string{"Jason"}, EnglishAdditionalTopSuggestions("JSON", nil)) // ToLower
	require.Equal(t, []string{"Wednesday"}, EnglishAdditionalTopSuggestions("wensday", nil))
}

func TestEnglishAdditionalTopSuggestions_YsToIes(t *testing.T) {
	// without isMisspelled, ys arm does not fire (needs !isMisspelled(suggestion))
	require.Nil(t, EnglishAdditionalTopSuggestions("babys", nil))
	// when dict accepts "babies"
	require.Equal(t, []string{"babies"}, EnglishAdditionalTopSuggestions("babys", func(w string) bool {
		return w != "babies"
	}))
}

func TestEnglishAdditionalTopSuggestions_WiredOnSpeller(t *testing.T) {
	r := NewAbstractEnglishSpellerRule("MORFOLOGIK_RULE_EN_US", "en-US", nil)
	require.NotNil(t, r.GetAdditionalTopSuggestionsFn)
	sp := morfologik.NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	sp.AddWord("ok")
	// Ths misspelled
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	// re-bind top fn with updated IsMisspelled
	r.GetAdditionalTopSuggestionsFn = func(existing []string, word string) []string {
		return EnglishAdditionalTopSuggestions(word, r.IsMisspelled)
	}
	m, err := r.Match(languagetool.AnalyzePlain("Ths"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	sugs := m[0].GetSuggestedReplacements()
	require.Contains(t, sugs, "This")
	require.Contains(t, sugs, "The")
	// curated should be first
	require.Equal(t, "This", sugs[0])
}
