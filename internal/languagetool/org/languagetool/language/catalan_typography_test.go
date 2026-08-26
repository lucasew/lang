package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCatalanAdvancedTypography(t *testing.T) {
	// Base Language: ellipsis
	require.Equal(t, "Això és…", CatalanAdvancedTypography("Això és..."))
	// Catalan quotes « »
	require.Equal(t, "Digues «hola»", CatalanAdvancedTypography(`Digues "hola"`))
	// PATTERN_1: l' → l’
	require.Equal(t, "l’home", CatalanAdvancedTypography("l'home"))
	// PATTERN_1 then base quotes: l'"home" → l’«home»
	require.Equal(t, "l’«home»", CatalanAdvancedTypography("l'\"home\""))
	require.True(t, CatalanIsAdvancedTypographyEnabled())
}

func TestCatalanRemoveOldDiacritics(t *testing.T) {
	require.Equal(t, "soc", CatalanRemoveOldDiacritics("sóc"))
	require.Equal(t, "dona", CatalanRemoveOldDiacritics("dóna"))
	require.Equal(t, "adeu", CatalanRemoveOldDiacritics("adéu"))
	require.Equal(t, "contrapel", CatalanRemoveOldDiacritics("contrapèl"))
	require.Equal(t, "Soc", CatalanRemoveOldDiacritics("Sóc"))
	// Java no-op pair for lowercase véns
	require.Equal(t, "véns", CatalanRemoveOldDiacritics("véns"))
	require.Equal(t, "Vens", CatalanRemoveOldDiacritics("Véns"))
	require.True(t, CatalanSuggestionNeedsOldDiacriticStrip("jo sóc aquí"))
	require.False(t, CatalanSuggestionNeedsOldDiacriticStrip("jo soc aquí"))
}

func TestFilterCatalanRuleMatches_OldDiacriticsOnSuggestions(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "X", Suggestions: []string{"sóc", "adeu"}},
	}
	out := FilterCatalanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, []string{"soc", "adeu"}, out[0].Suggestions)
}

func TestEnglishPrepareLineForSpeller(t *testing.T) {
	require.Equal(t, []string{"New York"}, EnglishPrepareLineForSpeller("New York\tNNP"))
	require.Equal(t, []string{"big"}, EnglishPrepareLineForSpeller("big\tJJ"))
	require.Equal(t, []string{""}, EnglishPrepareLineForSpeller("run\tVB"))
	require.Equal(t, []string{""}, EnglishPrepareLineForSpeller("foo+bar\tNN"))
	require.Equal(t, []string{"plain"}, EnglishPrepareLineForSpeller("plain"))
	require.Equal(t, []string{"word"}, EnglishPrepareLineForSpeller("word\tNN#comment"))
}
