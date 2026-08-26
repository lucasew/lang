package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestEnglishAddHyphenSuggestions(t *testing.T) {
	isMiss := func(w string) bool {
		return w != "well" && w != "known"
	}
	suggest := func(w string) []string {
		if w == "wel" {
			return []string{"well"}
		}
		if w == "knon" {
			return []string{"known"}
		}
		return nil
	}
	// one part wrong
	got := EnglishAddHyphenSuggestions([]string{"wel", "known"}, isMiss, suggest)
	require.Equal(t, []string{"well-known"}, got)
	// both parts wrong → two rebuilt suggestions
	got = EnglishAddHyphenSuggestions([]string{"wel", "knon"}, isMiss, suggest)
	require.Equal(t, []string{"well-knon", "wel-known"}, got)
	// no misspell parts
	got = EnglishAddHyphenSuggestions([]string{"well", "known"}, isMiss, suggest)
	require.Empty(t, got)
}

func TestHyphenatedWordSuggestion(t *testing.T) {
	require.Equal(t, "a-fixed-c", hyphenatedWordSuggestion([]string{"a", "b", "c"}, 1, "fixed"))
}

func TestEnglishHyphenSuggestions_WiredMatch(t *testing.T) {
	r := NewAbstractEnglishSpellerRule("MORFOLOGIK_RULE_EN_US", "en-US", nil)
	require.NotNil(t, r.AddHyphenSuggestionsFn)
	sp := morfologik.NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	sp.AddWord("well")
	sp.AddWord("known")
	// no suggestions map for whole "wel-known" → hyphen path
	sp.Suggestions["wel"] = []string{"well"}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	// re-bind hooks after Speller set
	r.GetAdditionalTopSuggestionsFn = func(existing []string, word string) []string {
		return EnglishAdditionalTopSuggestions(word, r.IsMisspelled)
	}
	r.AddHyphenSuggestionsFn = func(parts []string) []string {
		return EnglishAddHyphenSuggestions(parts, r.IsMisspelled, func(w string) []string {
			return r.Speller.FindReplacements(w)
		})
	}
	m, err := r.Match(languagetool.AnalyzePlain("wel-known"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "well-known")
}
