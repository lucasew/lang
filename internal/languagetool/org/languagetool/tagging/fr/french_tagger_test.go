package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestFrenchTagger(t *testing.T) {
	wt := tagging.MapWordTagger{"chien": {tagging.NewTaggedWord("chien", "N m s")}}
	got := NewFrenchTagger(wt).Tag([]string{"chien", "xyz"})
	require.Len(t, got, 2)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
}
