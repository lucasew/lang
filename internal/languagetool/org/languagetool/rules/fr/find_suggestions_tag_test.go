package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestFindSuggestionsFilter_WiresTagAndSynthesize(t *testing.T) {
	ClearFrenchFindSuggestionsTagger()
	f := NewFindSuggestionsFilter()
	require.NotNil(t, f.Tag)
	// Java FR overrides getSynthesizer() → FrenchSynthesizer.INSTANCE
	require.NotNil(t, f.Synthesize)
	require.Nil(t, f.Tag("maison")) // unwired fail-closed
}

func TestFilterTagWord_Wired(t *testing.T) {
	ClearFrenchFindSuggestionsTagger()
	t.Cleanup(ClearFrenchFindSuggestionsTagger)
	WireFrenchFindSuggestionsTagger(func(token string) []languagetool.TokenTag {
		if token == "maisons" {
			return []languagetool.TokenTag{{POS: "N f p", Lemma: "maison"}}
		}
		return nil
	})
	atr := FilterTagWord("maisons")
	require.NotNil(t, atr)
	require.True(t, atr.MatchesPosTagRegex("N.*"))
}
