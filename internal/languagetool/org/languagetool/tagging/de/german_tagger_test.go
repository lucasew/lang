package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestParseGermanPOS(t *testing.T) {
	a := ParseGermanPOS("SUB:NOM:SIN:MAS")
	require.Equal(t, POSNomen, a.Type)
	require.Equal(t, KasusNom, a.Kasus)
	require.Equal(t, NumerusSin, a.Numerus)
	require.Equal(t, GenusMas, a.Genus)
}

func TestGermanTagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"hund":  {tagging.NewTaggedWord("Hund", "SUB:NOM:SIN:MAS")},
		"Hunde": {tagging.NewTaggedWord("Hund", "SUB:NOM:PLU:MAS")},
	}
	tagger := NewGermanTagger(wt)
	got := tagger.Tag([]string{"Hunde", "xyz"})
	require.Len(t, got, 2)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
}

func TestVerbPrefixes(t *testing.T) {
	require.True(t, IsVerbPrefix("auf"))
	require.False(t, IsVerbPrefix("xyz"))
}
