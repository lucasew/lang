package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestPartizip2_NonSepPraes(t *testing.T) {
	// erstickt = er + stickt (VER:3:SIN:PRÄ:SFT)
	wt := tagging.MapWordTagger{
		"stickt": {tagging.NewTaggedWord("sticken", "VER:3:SIN:PRÄ:SFT")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"erstickt"})
	tags := posTagsOf(got[0])
	require.Contains(t, tags, "VER:PA2:SFT")
	require.Contains(t, tags, "PA2:PRD:GRU:VER")
}

// Twin: indexOf(word)==0 allows PA2 even when current idxPos > 0 (duplicate surface).
func TestPartizip2_IndexOfZeroNotIdxPos(t *testing.T) {
	wt := tagging.MapWordTagger{
		"stickt": {tagging.NewTaggedWord("sticken", "VER:3:SIN:PRÄ:SFT")},
	}
	tagger := NewGermanTagger(wt)
	// first "erstickt" at index 0; second should still get PA2 via indexOf==0
	got := tagger.Tag([]string{"erstickt", "und", "erstickt"})
	tags := posTagsOf(got[2])
	require.Contains(t, tags, "VER:PA2:SFT")
	require.Contains(t, tags, "PA2:PRD:GRU:VER")
}

func TestPartizip2_Declined(t *testing.T) {
	// erstickter = er + stickt + er
	wt := tagging.MapWordTagger{
		"stickt": {tagging.NewTaggedWord("sticken", "VER:3:SIN:PRÄ:SFT")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"erstickter"})
	tags := posTagsOf(got[0])
	found := false
	nomCount := 0
	for _, tg := range tags {
		if stringsHasPrefix(tg, "PA2:") && stringsHasSuffix(tg, ":VER") {
			found = true
		}
		if tg == "PA2:NOM:SIN:MAS:GRU:SOL:VER" {
			nomCount++
		}
	}
	require.True(t, found, "expected declined PA2 tags, got %v", tags)
	// Java postagsPartizipEndingEr lists each tag twice → two identical readings
	require.Equal(t, 2, nomCount, "EndingEr list is duplicated in Java")
	// lemma = word without suffix "er"
	for _, r := range got[0].GetReadings() {
		if r.GetPOSTag() != nil && *r.GetPOSTag() == "PA2:NOM:SIN:MAS:GRU:SOL:VER" {
			require.NotNil(t, r.GetLemma())
			require.Equal(t, "erstickt", *r.GetLemma())
			break
		}
	}
}

func TestSwissTagger_UsesSentenceContext(t *testing.T) {
	// imperative short form needs multi-token sentence
	wt := tagging.MapWordTagger{
		"gehe": {tagging.NewTaggedWord("gehen", "VER:IMP:SIN:SFT")},
	}
	tagger := NewSwissGermanTagger(wt)
	got := tagger.Tag([]string{"Geh", "bitte"})
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.True(t, stringsHasPrefix(*got[0].GetReadings()[0].GetPOSTag(), "VER:IMP:SIN"))
}

func TestSwissTagger_ssToEszett(t *testing.T) {
	wt := tagging.MapWordTagger{
		"Maß": {tagging.NewTaggedWord("Maß", "SUB:NOM:SIN:NEU")},
	}
	tagger := NewSwissGermanTagger(wt)
	got := tagger.Tag([]string{"Mass"})
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	// surface stays Swiss ss
	require.Equal(t, "Mass", got[0].GetReadings()[0].GetToken())
}
