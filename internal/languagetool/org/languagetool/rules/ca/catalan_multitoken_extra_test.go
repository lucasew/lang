package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"
	"github.com/stretchr/testify/require"
)

func TestCatalanMultitoken_ResourcePathsAndAdditionalHook(t *testing.T) {
	require.Equal(t, []string{
		"/ca/multiwords.txt",
		"/spelling_global.txt",
		"/ca/hyphenated_words.txt",
	}, CatalanMultitokenResourcePaths)
	require.NotNil(t, CatalanMultitokenSpellerInstance)
	s := NewCatalanMultitokenSpeller()
	require.NotNil(t, s.GetAdditionalSuggestions)
	// inject additional suggestions
	s.GetAdditionalSuggestions = func(w string) []multitoken.WeightedSuggestion {
		return []multitoken.WeightedSuggestion{{Word: "foo bar", Weight: 0}}
	}
	// without dict multiwords, additional still surfaces
	got := s.GetSuggestions("foo baz")
	require.Contains(t, got, "foo bar")
	// additional equals original → empty (Java)
	s.GetAdditionalSuggestions = func(w string) []multitoken.WeightedSuggestion {
		return []multitoken.WeightedSuggestion{{Word: w, Weight: 0}}
	}
	require.Empty(t, s.GetSuggestions("exact match"))
}

func TestCatalanMorfologikGetSpeller_DictMissing(t *testing.T) {
	ResetCatalanMultitokenSpeller()
	SetCatalanMultitokenDictExistsFn(func(path string) bool {
		require.Equal(t, CatalanSpellingMultitokenDict, path)
		return false
	})
	require.Nil(t, GetSpeller())
	ResetCatalanMultitokenSpeller()
}
