package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDiscoverGermanDisambiguationXML(t *testing.T) {
	p := DiscoverGermanDisambiguationXML()
	if p == "" {
		t.Skip("no DE disambiguation.xml in workspace")
	}
	require.Contains(t, p, "disambiguation.xml")
}

func TestWireGermanDisambiguator_NoPanic(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de")
	ok := WireGermanDisambiguator(lt)
	// May be false without resources; when true, Disambiguator is set.
	if ok {
		require.NotNil(t, lt.Disambiguator)
	}
	// nil tool → false
	require.False(t, WireGermanDisambiguator(nil))
}

func TestWireGermanRuntimeResourcesFor_Disambiguator(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de")
	require.NotPanics(t, func() {
		WireGermanRuntimeResourcesFor(lt, "DE")
	})
	// If disambiguation resources exist, Disambiguator should be set.
	if DiscoverGermanDisambiguationXML() != "" || DiscoverGermanResourceDir() != "" {
		// multitoken files alone are enough for a hybrid
		if DiscoverGermanDisambiguationXML() != "" {
			require.NotNil(t, lt.Disambiguator)
		}
	}
}
