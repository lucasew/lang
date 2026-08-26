package de

// Morph / unit helpers for AgreementSuggestor2 (algorithm pieces).
// Full Java matrix lives in AgreementSuggestor2Test_test.go.

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplaceSuggestorVars(t *testing.T) {
	got := replaceSuggestorVars(detTemplate, "SIN", "NEU", "AKK")
	// IND/DEF still present until getDetOrPronounSynth replaces it
	require.Contains(t, got, "SIN")
	require.Contains(t, got, "NEU")
	require.Contains(t, got, "AKK")
	require.NotContains(t, got, "PLU")
	require.NotContains(t, got, "MAS/FEM")
	require.Equal(t, "ART:IND/DEF:AKK:SIN:NEU", got)
}

func TestAgreementSuggestionEquals(t *testing.T) {
	// Java Suggestion.equals: phrase + tokenLevelEdits (not charLevelCorrections)
	a := agreementSuggestion{phrase: "dem Haus", tokenLevelEdits: 1, charLevelCorrections: 3}
	b := agreementSuggestion{phrase: "dem Haus", tokenLevelEdits: 1, charLevelCorrections: 9}
	c := agreementSuggestion{phrase: "dem Haus", tokenLevelEdits: 2, charLevelCorrections: 3}
	require.True(t, containsSuggestion([]agreementSuggestion{a}, b))
	require.False(t, containsSuggestion([]agreementSuggestion{a}, c))
}

func TestAgreementLevenshtein(t *testing.T) {
	require.Equal(t, 0, agreementLevenshtein("abc", "abc"))
	require.Equal(t, 1, agreementLevenshtein("abc", "ab"))
	require.Equal(t, 1, agreementLevenshtein("abc", "abd"))
	require.Equal(t, 3, agreementLevenshtein("", "abc"))
}

func TestApplyReplacementContractions_ZurIns(t *testing.T) {
	// Twin surfaces of addContraction
	got := applyReplacementContractions([]string{"der Frau", "dem Mann", "den Männern"}, ReplZur)
	require.Equal(t, []string{"zur Frau", "zum Mann", "zu Männern"}, got)

	got = applyReplacementContractions([]string{"das Haus", "dem Hause", "die Häuser", "den Häusern", "der Mann"}, ReplIns)
	require.Equal(t, []string{"ins Haus", "im Hause", "in die Häuser", "in den Häusern"}, got)
}
