package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatalanMorfologikMultitokenSpeller(t *testing.T) {
	require.Equal(t, "/ca/ca-ES_spelling_multitoken.dict", CatalanSpellingMultitokenDict)
	require.Nil(t, GetCatalanMultitokenSpellerSuggestions(nil, "x"))
	sugg := GetCatalanMultitokenSpellerSuggestions(func(path string) (func(string) []string, error) {
		require.Equal(t, CatalanSpellingMultitokenDict, path)
		return func(w string) []string { return []string{w + "!"} }, nil
	}, "foo")
	require.Equal(t, []string{"foo!"}, sugg)
}
