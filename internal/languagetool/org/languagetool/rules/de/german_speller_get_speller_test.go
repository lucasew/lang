package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestGetSpellingFilePaths_DE(t *testing.T) {
	// Java GermanSpellerRule.getSpellingFilePaths("de")
	require.Equal(t, []string{
		"/de/hunspell/spelling.txt",
		"/de/hunspell/spelling_custom.txt",
		"/de/multitoken-suggest.txt",
		"spelling_global.txt",
		"/de/hunspell/spelling_recommendation.txt",
	}, GetSpellingFilePaths("de"))
	// base CompoundAware without recommendation
	require.NotEqual(t, GetSpellingFilePaths("de"),
		[]string{
			"/de/hunspell/spelling.txt",
			"/de/hunspell/spelling_custom.txt",
			"/de/multitoken-suggest.txt",
			"spelling_global.txt",
		})
}

func TestGetSpeller_DE(t *testing.T) {
	if morfologik.DiscoverLanguageDict("/de/hunspell/de_DE.dict") == "" {
		t.Skip("de_DE.dict not in tree")
	}
	m := GetSpeller("DE", "", nil)
	require.NotNil(t, m)
	require.GreaterOrEqual(t, len(m.Spellers), 1)
	// binary accepts common German forms
	require.False(t, m.IsMisspelled("Haus"))
	require.False(t, m.IsMisspelled("Software"))
	// nonsense still misspelled
	require.True(t, m.IsMisspelled("sdadsadasxyz"))
	// edit distance is 2 for German getSpeller
	require.Equal(t, GermanSpellerMaxEditDistance, 2)
}

func TestGetSpeller_AT_variantPath(t *testing.T) {
	if morfologik.DiscoverLanguageDict("/de/hunspell/de_AT.dict") == "" &&
		morfologik.DiscoverLanguageDict("/de/hunspell/de_DE.dict") == "" {
		t.Skip("de dict not in tree")
	}
	// AT dict may be missing; GetSpeller returns null like Java resourceExists false
	m := GetSpeller("AT", AustrianGermanSpellingDict, nil)
	if morfologik.DiscoverLanguageDict("/de/hunspell/de_AT.dict") == "" {
		require.Nil(t, m)
		return
	}
	require.NotNil(t, m)
}

func TestGetSpeller_MissingBinary(t *testing.T) {
	// invent country code with no dict
	require.Nil(t, GetSpeller("XX", "", nil))
}

// InitFromDiscoveredResources wires GetSpeller into GermanSpellerRule.MorfoSpeller
// (Java memoized supplier from ctor → getSpeller).
func TestInitFromDiscoveredResources_WiresMorfoSpeller(t *testing.T) {
	if morfologik.DiscoverLanguageDict("/de/hunspell/de_DE.dict") == "" {
		t.Skip("de_DE.dict not in tree")
	}
	r := NewGermanSpellerRule(nil)
	require.NoError(t, r.InitFromDiscoveredResources())
	require.NotNil(t, r.MorfoSpeller, "getSpeller should wire when de_DE.dict present")
	// plain-text path should accept forms from spelling lists / multi when in multi
	require.False(t, r.MorfoSpeller.IsMisspelled("Haus"))
	// Suggest uses morfoSuggest → MorfoSpeller (not only FilterDict)
	sugs := r.Suggest("sdadsadasxyz")
	// nonsense may still get empty or unrelated; ensure no panic and path works
	_ = sugs
	// known typo path: FilterDict/morfo should yield something for close edits when available
	if sugs2 := r.Suggest("Huas"); len(sugs2) > 0 {
		// Haus is a likely suggestion
		require.Contains(t, sugs2, "Haus")
	}
}
