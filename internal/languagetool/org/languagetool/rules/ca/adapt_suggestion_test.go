package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testAdaptSuggestions (Catalan surface cases).
func TestAdaptSuggestion(t *testing.T) {
	require.Equal(t, "L'IEC", AdaptSuggestion("L'IEC", ""))
	require.Equal(t, "T'estimava", AdaptSuggestion("te estimava", "Estimava"))
	require.Equal(t, "l'Albert", AdaptSuggestion("el Albert", ""))
	require.Equal(t, "l'Albert", AdaptSuggestion("l'Albert", ""))
	require.Equal(t, "l'«Albert»", AdaptSuggestion("l'«Albert»", ""))
	require.Equal(t, "l’«Albert»", AdaptSuggestion("l’«Albert»", ""))
	require.Equal(t, `l'"Albert"`, AdaptSuggestion(`l'"Albert"`, ""))
	require.Equal(t, "em tancava", AdaptSuggestion("m'tancava", ""))
	require.Equal(t, "es tancava", AdaptSuggestion("s'tancava", ""))
	require.Equal(t, "l'R+D", AdaptSuggestion("l'R+D", ""))
	require.Equal(t, "l'FBI", AdaptSuggestion("l'FBI", ""))
}
