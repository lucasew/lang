package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func posTagsOf(rd *languagetool.AnalyzedTokenReadings) []string {
	var out []string
	if rd == nil {
		return out
	}
	for _, r := range rd.GetReadings() {
		if r != nil && r.GetPOSTag() != nil {
			out = append(out, *r.GetPOSTag())
		}
	}
	return out
}

func TestImpPraesSFT_Mutual_AtStart(t *testing.T) {
	wt := tagging.MapWordTagger{
		"habe": {tagging.NewTaggedWord("haben", "VER:IMP:SIN:SFT")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"Habe", "Zeit"})
	tags := posTagsOf(got[0])
	require.Contains(t, tags, "VER:IMP:SIN:SFT")
	require.Contains(t, tags, "VER:1:SIN:PRÄ:SFT")
}

func TestImpPraesSFT_Mutual_FromPraes(t *testing.T) {
	wt := tagging.MapWordTagger{
		"mache": {tagging.NewTaggedWord("machen", "VER:1:SIN:PRÄ:SFT")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"mache", "das"})
	tags := posTagsOf(got[0])
	require.Contains(t, tags, "VER:1:SIN:PRÄ:SFT")
	require.Contains(t, tags, "VER:IMP:SIN:SFT")
}

func TestImpPraesSFT_SkipSeparablePrefix(t *testing.T) {
	wt := tagging.MapWordTagger{
		"aufmachen": {tagging.NewTaggedWord("aufmachen", "VER:INF:NON")},
		"machen":    {tagging.NewTaggedWord("machen", "VER:IMP:SIN:SFT")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"aufmachen"})
	tags := posTagsOf(got[0])
	require.NotContains(t, tags, "VER:1:SIN:PRÄ:SFT")
}

func TestElativeUnknown(t *testing.T) {
	// lastPart must be length > 3 (Java lastPart.length() > 3)
	wt := tagging.MapWordTagger{
		"schön": {tagging.NewTaggedWord("schön", "ADJ:PRD:GRU")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"superschön"})
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "ADJ:PRD:GRU", *got[0].GetReadings()[0].GetPOSTag())
	require.NotNil(t, got[0].GetReadings()[0].GetLemma())
	require.Equal(t, "superschön", *got[0].GetReadings()[0].GetLemma())
}

func TestNonSeparablePrefix_Verzeih(t *testing.T) {
	// verzeih → base zeih with IMP → mutual on surface
	wt := tagging.MapWordTagger{
		"verzeih": {tagging.NewTaggedWord("verzeihen", "VER:IMP:SIN:SFT")},
		"zeih":    {tagging.NewTaggedWord("zeihen", "VER:IMP:SIN:SFT")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"Verzeih", "mir"})
	tags := posTagsOf(got[0])
	require.Contains(t, tags, "VER:IMP:SIN:SFT")
	require.Contains(t, tags, "VER:1:SIN:PRÄ:SFT")
}
