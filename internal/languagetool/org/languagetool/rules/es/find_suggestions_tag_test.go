package es

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestFindSuggestionsFilter_WiresTagNoSynthesize(t *testing.T) {
	ClearSpanishFindSuggestionsTagger()
	f := NewFindSuggestionsFilter()
	require.NotNil(t, f.Tag)
	// Java ES FindSuggestionsFilter does not override getSynthesizer()
	require.Nil(t, f.Synthesize)
	require.Nil(t, f.Tag("casa"))
}

func TestFilterTagWord_Wired(t *testing.T) {
	ClearSpanishFindSuggestionsTagger()
	t.Cleanup(ClearSpanishFindSuggestionsTagger)
	WireSpanishFilterTaggerFromTagWord(func(token string) []languagetool.TokenTag {
		if token == "casas" {
			return []languagetool.TokenTag{{POS: "NCFP000", Lemma: "casa"}}
		}
		return nil
	})
	atr := FilterTagWord("casas")
	require.NotNil(t, atr)
	require.True(t, atr.MatchesPosTagRegex("NC.*"))
}
