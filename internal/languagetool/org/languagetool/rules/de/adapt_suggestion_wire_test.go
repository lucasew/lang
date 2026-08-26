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
	f := WireAdaptSuggestionFilter()
	require.NotNil(t, f)
	if f.GenderOf != nil {
		_ = f.GenderOf("Haus")
	}
	_ = f.Synthesize
	require.NotNil(t, f.FilterSuggestions, "AgreementRule SuggestionFilter must be wired")
}

func TestNounGenderFromTagger_Manual(t *testing.T) {
	root := DiscoverGermanResourceDir()
	if root == "" {
		t.Skip("no DE resources")
	}
	tagger := openDiscoveredGermanTagger(root)
	if tagger == nil {
		t.Skip("no tagger sources")
	}
	_ = nounGenderFromTagger(tagger, "vorm")
}
