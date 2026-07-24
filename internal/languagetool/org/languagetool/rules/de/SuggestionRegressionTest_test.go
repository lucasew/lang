package de

// Twin of SuggestionRegressionTest (Java interactive / no unit @Test in CI).
// Soft AdaptSuggestion surface for DE spelling suggestions.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/hunspell"
	"github.com/stretchr/testify/require"
)

// Port of SuggestionRegressionTest (no @Test)
func TestSuggestionRegression_NoTests(t *testing.T) {
	// inject dict: regression-style misspelling → suggestion
	dict := hunspell.NewMapHunspellDictionary([]string{"Haus", "Hund", "Straße", "Strasse"})
	dict.SetSuggestions("Huas", []string{"Haus"})
	dict.SetSuggestions("Strase", []string{"Straße", "Strasse"})
	r := hunspell.NewHunspellRule("de-DE", dict)

	m, err := r.Match(languagetool.AnalyzePlain("Huas"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "Haus")
	// adapt suggestion identity soft
	require.Equal(t, "Haus", languagetool.AdaptSuggestion("Haus", "Huas"))
}
