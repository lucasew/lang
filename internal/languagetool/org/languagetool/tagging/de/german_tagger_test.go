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

// Twin AnalyzedGermanToken constructor behaviors.
func TestParseGermanPOS_ShortTagNullFields(t *testing.T) {
	a := ParseGermanPOS("ADV")
	require.Equal(t, POSType(""), a.Type)
	require.Equal(t, Kasus(""), a.Kasus)
	a2 := ParseGermanPOS("VER:INF")
	// only 2 parts after split? "VER:INF" → ["VER","INF"] length 2 < 3
	require.Equal(t, POSType(""), a2.Type)
}

func TestParseGermanPOS_EIGOverSUB(t *testing.T) {
	a := ParseGermanPOS("EIG:NOM:SIN:MAS")
	require.Equal(t, POSProperNoun, a.Type)
	// SUB after EIG does not override
	a2 := ParseGermanPOS("SUB:EIG:NOM:SIN:MAS")
	// EIG always assigns even after SUB set
	require.Equal(t, POSProperNoun, a2.Type)
}

func TestParseGermanPOS_PAOverSUB(t *testing.T) {
	// PA always assigns
	a := ParseGermanPOS("SUB:PA2:NOM:SIN:MAS")
	require.Equal(t, POSPartizip, a.Type)
}

func TestParseGermanPOS_NOGMapsToFem(t *testing.T) {
	// Java: NOG → Genus.FEMININUM
	a := ParseGermanPOS("SUB:NOM:PLU:NOG")
	require.Equal(t, POSNomen, a.Type)
	require.Equal(t, GenusFem, a.Genus)
	require.Equal(t, NumerusPlu, a.Numerus)
}

func TestParseGermanPOS_Determination(t *testing.T) {
	a := ParseGermanPOS("ART:DEF:NOM:SIN:FEM")
	require.Equal(t, POSDeterminer, a.Type)
	require.Equal(t, DetDefinite, a.Determination)
	require.Equal(t, KasusNom, a.Kasus)
	require.Equal(t, GenusFem, a.Genus)
}
