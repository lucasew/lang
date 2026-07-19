package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestUkrainianTagger_DynamicTaggingPiv(t *testing.T) {
	wt := tagging.MapWordTagger{
		"години": {tagging.NewTaggedWord("година", "noun:inanim:p:v_naz")},
	}
	tg := NewUkrainianTagger(wt)
	got := tg.Tag([]string{"півгодини"})
	require.True(t, got[0].IsTagged(), "пів+години should tag via no-dash prefix")
	require.Contains(t, *got[0].GetReadings()[0].GetLemma(), "годин")
}

func TestUkrainianTagger_DynamicTaggingPrefixes(t *testing.T) {
	wt := tagging.MapWordTagger{
		"тест": {tagging.NewTaggedWord("тест", "noun")},
	}
	tg := NewUkrainianTagger(wt)
	got := tg.Tag([]string{"супертест"})
	require.True(t, got[0].IsTagged())
}

func TestDashPrefixesOfficial(t *testing.T) {
	// invent subset lists are gone — official dash_prefixes.txt
	require.True(t, IsDashPrefix("міні") || IsDashPrefix("анти") || IsDashPrefix("AI"),
		"expected common prefixes from dash_prefixes.txt")
	// no-dash list from invalid + alt keys
	list := NoDashPrefixList()
	require.NotEmpty(t, list)
	// longest-first: напів before пів if both present
	hasNapiv, hasPiv := false, false
	for _, p := range list {
		if p == "напів" {
			hasNapiv = true
		}
		if p == "пів" {
			hasPiv = true
		}
	}
	if hasNapiv && hasPiv {
		// ensure напів appears before пів in list
		iNapiv, iPiv := -1, -1
		for i, p := range list {
			if p == "напів" {
				iNapiv = i
			}
			if p == "пів" {
				iPiv = i
			}
		}
		require.Less(t, iNapiv, iPiv, "longest-first order")
	}
}
