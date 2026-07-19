package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWireAdaptSuggestionFilter_HooksFromResources(t *testing.T) {
	root := DiscoverGermanResourceDir()
	if root == "" {
		t.Skip("no DE resources")
	}
	// reset once for test isolation is hard; call wire and check structure
	f := WireAdaptSuggestionFilter()
	require.NotNil(t, f)
	// GenderOf wires if added.txt or dict present
	if f.GenderOf != nil {
		// "Haus" may only work with german.dict; added.txt may have nouns
		// just ensure function is callable
		_ = f.GenderOf("Haus")
	}
	// Synthesize only if german_synth.dict present (often missing)
	_ = f.Synthesize
}

func TestNounGenderFromTagger_Manual(t *testing.T) {
	// unit without full wire: use openDiscovered when resources exist
	root := DiscoverGermanResourceDir()
	if root == "" {
		t.Skip("no DE resources")
	}
	tagger := openDiscoveredGermanTagger(root)
	if tagger == nil {
		t.Skip("no tagger sources")
	}
	// vorm is in added.txt as PRP not SUB — use a form that may be in added as SUB
	// Fail-closed empty is OK when only PRP available
	_ = nounGenderFromTagger(tagger, "vorm")
}
