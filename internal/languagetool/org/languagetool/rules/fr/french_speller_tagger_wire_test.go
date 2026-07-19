package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	taggingfr "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/fr"
	"github.com/stretchr/testify/require"
)

func TestWireFrenchSpellerTagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"arrive": {tagging.NewTaggedWord("arriver", "V ind pres 3 s")},
	}
	tg := taggingfr.NewFrenchTagger(wt)
	r := NewMorfologikFrenchSpellerRule()
	WireFrenchSpellerTagger(r, tg)
	require.NotNil(t, r.TagPOS)
	require.Equal(t, []string{"V ind pres 3 s"}, r.TagPOS("arrive"))
	// apostrophe path uses TagPOS
	got := r.apostropheHyphenTopSuggestions("larrive")
	require.Equal(t, []string{"l'arrive"}, got)
}

func TestWireFrenchSpellerTagPOS(t *testing.T) {
	r := NewMorfologikFrenchSpellerRule()
	WireFrenchSpellerTagPOS(r, func(token string) []languagetool.TokenTag {
		if token == "maison" {
			return []languagetool.TokenTag{{POS: "N f s"}}
		}
		return nil
	})
	require.True(t, r.isTagged("maison"))
	require.Equal(t, "maison 2", r.digitSplitTopSuggestion("maison2"))
}
