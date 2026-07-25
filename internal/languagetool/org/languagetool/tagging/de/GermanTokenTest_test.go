package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of GermanToken nested type toString() values (Java object identity via name string).
// Java: GermanToken.POSType / Kasus / Numerus / Genus / Determination each store a name
// returned by toString(); no other public API.

func TestGermanToken_POSType_ToString(t *testing.T) {
	// Java: new POSType("…") / public static finals
	require.Equal(t, "Nomen", string(POSNomen))
	require.Equal(t, "Verb", string(POSVerb))
	require.Equal(t, "Adjektiv", string(POSAdjektiv))
	require.Equal(t, "Determiner", string(POSDeterminer))
	require.Equal(t, "Pronomen", string(POSPronomen))
	require.Equal(t, "Partizip", string(POSPartizip))
	require.Equal(t, "Eigenname", string(POSProperNoun))
	require.Equal(t, "Other", string(POSOther))
}

func TestGermanToken_Kasus_ToString(t *testing.T) {
	require.Equal(t, "Nominativ", string(KasusNominativ))
	require.Equal(t, "Akkusativ", string(KasusAkkusativ))
	require.Equal(t, "Dativ", string(KasusDativ))
	require.Equal(t, "Genitiv", string(KasusGenitiv))
	require.Equal(t, "Other", string(KasusOther))
	// Morphy-oriented aliases must share the Java toString values (not Morphy codes).
	require.Equal(t, KasusNominativ, KasusNom)
	require.Equal(t, KasusAkkusativ, KasusAkk)
	require.Equal(t, KasusDativ, KasusDat)
	require.Equal(t, KasusGenitiv, KasusGen)
}

func TestGermanToken_Numerus_ToString(t *testing.T) {
	require.Equal(t, "Singular", string(NumerusSingular))
	require.Equal(t, "Plural", string(NumerusPlural))
	require.Equal(t, "Other", string(NumerusOther))
	require.Equal(t, NumerusSingular, NumerusSin)
	require.Equal(t, NumerusPlural, NumerusPlu)
}

func TestGermanToken_Genus_ToString(t *testing.T) {
	// Java declaration order: NEUTRUM, MASKULINUM, FEMININUM, OTHER, ALLGEMEIN
	require.Equal(t, "Neutrum", string(GenusNeutrum))
	require.Equal(t, "Maskulinum", string(GenusMaskulinum))
	require.Equal(t, "Femininum", string(GenusFemininum))
	require.Equal(t, "Other", string(GenusOther))
	require.Equal(t, "Allgemein", string(GenusAllgemein))
	require.Equal(t, GenusMaskulinum, GenusMas)
	require.Equal(t, GenusFemininum, GenusFem)
	require.Equal(t, GenusNeutrum, GenusNeu)
	require.Equal(t, GenusAllgemein, GenusALG)
}

func TestGermanToken_Determination_ToString(t *testing.T) {
	// Java: "definit" / "indefinit" (lowercase)
	require.Equal(t, "definit", string(DetDefinite))
	require.Equal(t, "indefinit", string(DetIndefinite))
	require.Equal(t, DetDefinite, DeterminationDefinite)
	require.Equal(t, DetIndefinite, DeterminationIndefinite)
}

// ParseGermanPOS maps Morphy codes → GermanToken objects (Java AnalyzedGermanToken ctor).
// Identity comparisons must match the Java-visible enum objects.
func TestParseGermanPOS_MapsMorphyToGermanTokenObjects(t *testing.T) {
	a := ParseGermanPOS("ART:DEF:NOM:SIN:FEM")
	require.Equal(t, POSDeterminer, a.Type)
	require.Equal(t, KasusNominativ, a.Kasus)
	require.Equal(t, NumerusSingular, a.Numerus)
	require.Equal(t, GenusFemininum, a.Genus)
	require.Equal(t, DetDefinite, a.Determination)
	// toString forms used by AgreementTools.makeString
	require.Equal(t, "Nominativ", string(a.Kasus))
	require.Equal(t, "Singular", string(a.Numerus))
	require.Equal(t, "Femininum", string(a.Genus))
	require.Equal(t, "definit", string(a.Determination))

	alg := ParseGermanPOS("PRO:IND:NOM:PLU:ALG")
	require.Equal(t, GenusAllgemein, alg.Genus)
	require.Equal(t, "Allgemein", string(alg.Genus))

	nog := ParseGermanPOS("SUB:NOM:PLU:NOG")
	// Java: NOG → Genus.FEMININUM
	require.Equal(t, GenusFemininum, nog.Genus)
}
