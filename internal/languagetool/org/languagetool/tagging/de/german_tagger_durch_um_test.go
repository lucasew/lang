package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestDurch_DualPA2(t *testing.T) {
	// durchläuft: durch + läuft (VER:3:SIN:PRÄ)
	wt := tagging.MapWordTagger{
		"läuft": {tagging.NewTaggedWord("laufen", "VER:3:SIN:PRÄ:SFT")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"durchläuft"})
	tags := posTagsOf(got[0])
	require.Contains(t, tags, "VER:3:SIN:PRÄ:SFT:NEB")
	require.Contains(t, tags, "VER:PA2:SFT")
	require.Contains(t, tags, "PA2:PRD:GRU:VER")
}

func TestUm_DualPA2(t *testing.T) {
	wt := tagging.MapWordTagger{
		"fährt": {tagging.NewTaggedWord("fahren", "VER:3:SIN:PRÄ:NON")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"umfährt"})
	tags := posTagsOf(got[0])
	require.Contains(t, tags, "VER:PA2:SFT")
	require.Contains(t, tags, "PA2:PRD:GRU:VER")
}
