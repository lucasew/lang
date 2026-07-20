package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestFindSuggestionsFilter_WiresTagAndSynthesize(t *testing.T) {
	ClearCatalanFindSuggestionsTagger()
	f := NewFindSuggestionsFilter()
	require.NotNil(t, f.Tag)
	require.NotNil(t, f.Synthesize)
	require.Nil(t, f.Tag("casa"))
}

func TestFilterTagWord_Wired(t *testing.T) {
	ClearCatalanFindSuggestionsTagger()
	t.Cleanup(ClearCatalanFindSuggestionsTagger)
	WireCatalanFilterTaggerFromTagWord(func(token string) []languagetool.TokenTag {
		if token == "cases" {
			return []languagetool.TokenTag{{POS: "NCFP000", Lemma: "casa"}}
		}
		return nil
	})
	atr := FilterTagWord("cases")
	require.NotNil(t, atr)
	require.True(t, atr.MatchesPosTagRegex("NC.*"))
}
