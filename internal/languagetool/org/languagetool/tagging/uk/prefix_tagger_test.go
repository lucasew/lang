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

