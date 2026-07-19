package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLookupAdditionalRegexp_Samples(t *testing.T) {
	require.Greater(t, len(additionalSugRepls), 100)
	require.Greater(t, len(additionalSugPatternFixed), 50)
	require.Equal(t, []string{"handytipps"}, lookupAdditionalSuggestionsRegexp("handytips"))
	require.Equal(t, []string{"widerstehen"}, lookupAdditionalSuggestionsRegexp("wiederstehen"))
	require.Equal(t, []string{"Ghanaer"}, lookupAdditionalSuggestionsRegexp("Ghanesen"))
	require.Equal(t, []string{"upgedatet"}, lookupAdditionalSuggestionsRegexp("geupdated"))
	r := NewGermanSpellerRule(nil)
	require.Equal(t, []string{"widerstehen"}, r.Suggest("wiederstehen"))
	require.Equal(t, []string{"dass"}, r.Suggest("daß"))
}

func TestLookupAdditionalLambda_Samples(t *testing.T) {
	// [mM]illion(en)?mal → " Mal"
	require.Equal(t, []string{"Million Mal"}, lookupAdditionalSuggestionsLambda("Millionmal"))
	// uppercaseFirstChar after "mal"→" Mal"
	require.Equal(t, []string{"Million Mal"}, lookupAdditionalSuggestionsLambda("millionmal"))
	// problemhaft
	got := lookupAdditionalSuggestionsLambda("problemhaft")
	require.Contains(t, got, "problembehaftet")
	require.Contains(t, got, "problematisch")
	// Panelen
	got2 := lookupAdditionalSuggestionsLambda("Panelen")
	require.Contains(t, got2, "Paneelen")
	require.Contains(t, got2, "Panels")
	// via Suggest
	r := NewGermanSpellerRule(nil)
	require.Equal(t, []string{"Million Mal"}, r.Suggest("Millionmal"))
}

func TestAbbreviationSuggestion(t *testing.T) {
	r := NewGermanSpellerRule(nil)
	require.Nil(t, r.abbreviationSuggestion("usw")) // no tagger
	r.TagPOS = func(w string) []string {
		if w == "usw" {
			return []string{"ABK:ADV"}
		}
		return nil
	}
	require.Equal(t, []string{"usw."}, r.abbreviationSuggestion("usw"))
	require.Nil(t, r.abbreviationSuggestion("abcde")) // length >= 5
}

func TestSuggestHyphenatedCompound(t *testing.T) {
	ClearGermanFilterSpeller()
	t.Cleanup(ClearGermanFilterSpeller)
	dict := deHunspellPath(t, "de_DE.dict")
	if !WireGermanFilterSpeller(dict) {
		t.Skip("no dict")
	}
	r := NewGermanSpellerRule(nil)
	// Netflix-Flm style: second part misspelled if Flm not in dict
	sugs := r.suggestHyphenatedCompound("Netflix-Flm")
	t.Logf("Netflix-Flm -> %v", sugs)
	// if dict can suggest for Flm, non-empty
}

func TestInitWiresCompoundTokenize(t *testing.T) {
	if DiscoverGermanResourceDir() == "" {
		t.Skip("no resources")
	}
	r := NewGermanSpellerRule(nil)
	require.NoError(t, r.InitFromDiscoveredResources())
	require.NotNil(t, r.CompoundTokenize)
	require.NotNil(t, r.CompoundTokenizeNonStrict)
	// Autobahn-like if auto+bahn in common words
	parts := r.CompoundTokenize("Autobahn")
	t.Logf("Autobahn -> %v", parts)
}
