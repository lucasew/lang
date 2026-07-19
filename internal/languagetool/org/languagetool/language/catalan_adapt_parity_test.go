package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Java JLanguageToolTest.testAdaptSuggestions (Catalan) — same cases as rules/ca.
func TestCatalanAdaptSuggestion_JavaParity(t *testing.T) {
	require.Equal(t, "L'IEC", CatalanAdaptSuggestion("L'IEC", ""))
	require.Equal(t, "T'estimava", CatalanAdaptSuggestion("te estimava", "Estimava"))
	require.Equal(t, "l'Albert", CatalanAdaptSuggestion("el Albert", ""))
	require.Equal(t, "l'Albert", CatalanAdaptSuggestion("l'Albert", ""))
	require.Equal(t, "l'«Albert»", CatalanAdaptSuggestion("l'«Albert»", ""))
	require.Equal(t, "l’«Albert»", CatalanAdaptSuggestion("l’«Albert»", ""))
	require.Equal(t, `l'"Albert"`, CatalanAdaptSuggestion(`l'"Albert"`, ""))
	require.Equal(t, "em tancava", CatalanAdaptSuggestion("m'tancava", ""))
	require.Equal(t, "es tancava", CatalanAdaptSuggestion("s'tancava", ""))
	require.Equal(t, "l'R+D", CatalanAdaptSuggestion("l'R+D", ""))
	require.Equal(t, "l'FBI", CatalanAdaptSuggestion("l'FBI", ""))
}
