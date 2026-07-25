package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Twin of org.languagetool.tagging.de.AnalyzedGermanToken constructor + getters.
// Java: AnalyzedGermanToken(AnalyzedToken) parses token.getPOSTag() via StringUtils.split(pos, ':').

func analyzedFromPOS(pos *string) AnalyzedGermanToken {
	return NewAnalyzedGermanToken(languagetool.NewAnalyzedToken("x", pos, nil))
}

func ptr(s string) *string { return &s }

func TestAnalyzedGermanToken_NullOrShortPOS_AllFieldsNull(t *testing.T) {
	// Java: posTag == null → all fields null
	a := NewAnalyzedGermanToken(languagetool.NewAnalyzedToken("x", nil, nil))
	require.Equal(t, POSType(""), a.GetType())
	require.Equal(t, Kasus(""), a.GetCasus())
	require.Equal(t, Numerus(""), a.GetNumerus())
	require.Equal(t, Genus(""), a.GetGenus())
	require.Equal(t, Determination(""), a.GetDetermination())

	// empty string: StringUtils.split("", ':').length == 0 < 3
	a = ParseGermanPOS("")
	require.Equal(t, POSType(""), a.GetType())
	require.Equal(t, Kasus(""), a.GetCasus())
	require.Equal(t, Numerus(""), a.GetNumerus())
	require.Equal(t, Genus(""), a.GetGenus())
	require.Equal(t, Determination(""), a.GetDetermination())

	// 1 part
	a = ParseGermanPOS("ADV")
	require.Equal(t, POSType(""), a.GetType())
	require.Equal(t, Kasus(""), a.GetCasus())
	require.Equal(t, Numerus(""), a.GetNumerus())
	require.Equal(t, Genus(""), a.GetGenus())
	require.Equal(t, Determination(""), a.GetDetermination())

	// 2 parts
	a = ParseGermanPOS("VER:INF")
	require.Equal(t, POSType(""), a.GetType())
	require.Equal(t, Kasus(""), a.GetCasus())
	require.Equal(t, Numerus(""), a.GetNumerus())
	require.Equal(t, Genus(""), a.GetGenus())
	require.Equal(t, Determination(""), a.GetDetermination())
}

func TestAnalyzedGermanToken_ART_DEF_NOM_SIN_FEM(t *testing.T) {
	// Java: ART:DEF:NOM:SIN:FEM
	a := analyzedFromPOS(ptr("ART:DEF:NOM:SIN:FEM"))
	require.Equal(t, POSDeterminer, a.GetType())
	require.Equal(t, KasusNominativ, a.GetCasus())
	require.Equal(t, NumerusSingular, a.GetNumerus())
	require.Equal(t, GenusFemininum, a.GetGenus())
	require.Equal(t, DetDefinite, a.GetDetermination())
	// toString forms (GermanToken nested types)
	require.Equal(t, "Determiner", string(a.GetType()))
	require.Equal(t, "Nominativ", string(a.GetCasus()))
	require.Equal(t, "Singular", string(a.GetNumerus()))
	require.Equal(t, "Femininum", string(a.GetGenus()))
	require.Equal(t, "definit", string(a.GetDetermination()))
}

func TestAnalyzedGermanToken_SUB_Basic(t *testing.T) {
	a := ParseGermanPOS("SUB:NOM:SIN:MAS")
	require.Equal(t, POSNomen, a.GetType())
	require.Equal(t, KasusNominativ, a.GetCasus())
	require.Equal(t, NumerusSingular, a.GetNumerus())
	require.Equal(t, GenusMaskulinum, a.GetGenus())
	require.Equal(t, Determination(""), a.GetDetermination())
	require.Equal(t, "Nomen", string(a.GetType()))
	require.Equal(t, "Maskulinum", string(a.GetGenus()))
}

func TestAnalyzedGermanToken_EIG_PriorityOverSUB(t *testing.T) {
	// EIG alone (≥3 parts)
	a := ParseGermanPOS("EIG:NOM:SIN:MAS")
	require.Equal(t, POSProperNoun, a.GetType())
	require.Equal(t, "Eigenname", string(a.GetType()))

	// SUB first then EIG: EIG always assigns → PROPER_NOUN
	a = ParseGermanPOS("SUB:EIG:NOM:SIN:MAS")
	require.Equal(t, POSProperNoun, a.GetType())

	// EIG first then SUB: SUB only if tempType still null → stays PROPER_NOUN
	a = ParseGermanPOS("EIG:SUB:NOM:SIN:MAS")
	require.Equal(t, POSProperNoun, a.GetType())
}

func TestAnalyzedGermanToken_PA1_PA2_Overwrite(t *testing.T) {
	// PA2 always assigns PARTIZIP even after SUB set type
	a := ParseGermanPOS("SUB:PA2:NOM:SIN:MAS")
	require.Equal(t, POSPartizip, a.GetType())
	require.Equal(t, "Partizip", string(a.GetType()))
	require.Equal(t, KasusNominativ, a.GetCasus())
	require.Equal(t, NumerusSingular, a.GetNumerus())
	require.Equal(t, GenusMaskulinum, a.GetGenus())

	// PA1 same
	a = ParseGermanPOS("SUB:PA1:AKK:PLU:NEU")
	require.Equal(t, POSPartizip, a.GetType())
	require.Equal(t, KasusAkkusativ, a.GetCasus())
	require.Equal(t, NumerusPlural, a.GetNumerus())
	require.Equal(t, GenusNeutrum, a.GetGenus())

	// PA after EIG overwrites (both always-assign)
	a = ParseGermanPOS("EIG:PA2:NOM:SIN:MAS")
	require.Equal(t, POSPartizip, a.GetType())

	// EIG after PA overwrites back to PROPER_NOUN
	a = ParseGermanPOS("PA2:EIG:NOM:SIN:MAS")
	require.Equal(t, POSProperNoun, a.GetType())
}

func TestAnalyzedGermanToken_VER_ADJ_PRO(t *testing.T) {
	// Need ≥3 parts for parse to run
	a := ParseGermanPOS("VER:1:SIN")
	require.Equal(t, POSVerb, a.GetType())
	require.Equal(t, NumerusSingular, a.GetNumerus())

	a = ParseGermanPOS("ADJ:PRD:GRU")
	require.Equal(t, POSAdjektiv, a.GetType())

	a = ParseGermanPOS("PRO:PER:NOM:SIN:MAS")
	require.Equal(t, POSPronomen, a.GetType())
	require.Equal(t, KasusNominativ, a.GetCasus())
	require.Equal(t, NumerusSingular, a.GetNumerus())
	require.Equal(t, GenusMaskulinum, a.GetGenus())

	// VER only if tempType still null — SUB first wins for type
	a = ParseGermanPOS("SUB:VER:NOM:SIN:MAS")
	require.Equal(t, POSNomen, a.GetType())
}

func TestAnalyzedGermanToken_NOG_MapsToFemininum(t *testing.T) {
	// Java: NOG → Genus.FEMININUM (comment: no genus because only used as plural)
	a := ParseGermanPOS("SUB:NOM:PLU:NOG")
	require.Equal(t, POSNomen, a.GetType())
	require.Equal(t, KasusNominativ, a.GetCasus())
	require.Equal(t, NumerusPlural, a.GetNumerus())
	require.Equal(t, GenusFemininum, a.GetGenus())
	require.Equal(t, "Femininum", string(a.GetGenus()))
}

func TestAnalyzedGermanToken_ALG_MapsToAllgemein(t *testing.T) {
	a := ParseGermanPOS("PRO:IND:NOM:PLU:ALG")
	require.Equal(t, POSPronomen, a.GetType())
	require.Equal(t, DetIndefinite, a.GetDetermination())
	require.Equal(t, KasusNominativ, a.GetCasus())
	require.Equal(t, NumerusPlural, a.GetNumerus())
	require.Equal(t, GenusAllgemein, a.GetGenus())
	require.Equal(t, "Allgemein", string(a.GetGenus()))
	require.Equal(t, "indefinit", string(a.GetDetermination()))
}

func TestAnalyzedGermanToken_Determination_IND_DEF(t *testing.T) {
	a := ParseGermanPOS("ART:IND:DAT:SIN:NEU")
	require.Equal(t, POSDeterminer, a.GetType())
	require.Equal(t, DetIndefinite, a.GetDetermination())
	require.Equal(t, KasusDativ, a.GetCasus())
	require.Equal(t, NumerusSingular, a.GetNumerus())
	require.Equal(t, GenusNeutrum, a.GetGenus())

	a = ParseGermanPOS("ART:DEF:GEN:PLU:MAS")
	require.Equal(t, DetDefinite, a.GetDetermination())
	require.Equal(t, KasusGenitiv, a.GetCasus())
	require.Equal(t, NumerusPlural, a.GetNumerus())
	require.Equal(t, GenusMaskulinum, a.GetGenus())
}

func TestAnalyzedGermanToken_Kasus_LastWins(t *testing.T) {
	// Java: each matching part assigns tempCasus (no first-wins guard) → last match wins.
	// Artificial multi-case tag documents that overwrite semantics.
	a := ParseGermanPOS("SUB:NOM:AKK:DAT:GEN")
	require.Equal(t, POSNomen, a.GetType())
	require.Equal(t, KasusGenitiv, a.GetCasus(), "last of NOM/AKK/DAT/GEN wins")
	require.Equal(t, "Genitiv", string(a.GetCasus()))

	a = ParseGermanPOS("SUB:GEN:NOM:SIN:MAS")
	require.Equal(t, KasusNominativ, a.GetCasus())
}

func TestAnalyzedGermanToken_Numerus_Genus_LastWins(t *testing.T) {
	a := ParseGermanPOS("SUB:NOM:SIN:PLU:MAS")
	require.Equal(t, NumerusPlural, a.GetNumerus())

	a = ParseGermanPOS("SUB:NOM:SIN:MAS:FEM")
	require.Equal(t, GenusFemininum, a.GetGenus())

	a = ParseGermanPOS("SUB:NOM:SIN:MAS:ALG")
	require.Equal(t, GenusAllgemein, a.GetGenus())

	a = ParseGermanPOS("ART:IND:DEF:NOM:SIN:MAS")
	require.Equal(t, DetDefinite, a.GetDetermination())
}

func TestAnalyzedGermanToken_StringUtilsSplit_EmptySegments(t *testing.T) {
	// Apache StringUtils.split discards empty tokens (preserveAllTokens=false).
	// ":A:B:C" → ["A","B","C"] length 3 → parse runs (not short).
	a := ParseGermanPOS(":ART:DEF:NOM")
	// parts after discard: ART, DEF, NOM → length 3
	require.Equal(t, POSDeterminer, a.GetType())
	require.Equal(t, DetDefinite, a.GetDetermination())
	require.Equal(t, KasusNominativ, a.GetCasus())

	// "A::B::C" → ["A","B","C"] length 3
	a = ParseGermanPOS("SUB::NOM::MAS")
	require.Equal(t, POSNomen, a.GetType())
	require.Equal(t, KasusNominativ, a.GetCasus())
	require.Equal(t, GenusMaskulinum, a.GetGenus())

	// "A:B:" → ["A","B"] length 2 < 3 → all null
	a = ParseGermanPOS("SUB:NOM:")
	require.Equal(t, POSType(""), a.GetType())
	require.Equal(t, Kasus(""), a.GetCasus())

	// ":::" → [] length 0 < 3
	a = ParseGermanPOS(":::")
	require.Equal(t, POSType(""), a.GetType())
}

func TestStringUtilsSplitColon(t *testing.T) {
	require.Nil(t, stringUtilsSplitColon(""))
	require.Equal(t, []string{"a"}, stringUtilsSplitColon("a"))
	require.Equal(t, []string{"a", "b", "c"}, stringUtilsSplitColon("a:b:c"))
	require.Equal(t, []string{"a", "b", "c"}, stringUtilsSplitColon(":a:b:c"))
	require.Equal(t, []string{"a", "b", "c"}, stringUtilsSplitColon("a:b:c:"))
	require.Equal(t, []string{"a", "b"}, stringUtilsSplitColon("a::b"))
	require.Equal(t, []string(nil), stringUtilsSplitColon(":::"))
	// empty result may be nil or empty slice; length is what matters for < 3 check
	require.Less(t, len(stringUtilsSplitColon(":::")), 3)
}

func TestAnalyzedGermanToken_GettersMatchFields(t *testing.T) {
	a := ParseGermanPOS("ART:DEF:DAT:PLU:NEU")
	require.Equal(t, a.Type, a.GetType())
	require.Equal(t, a.Kasus, a.GetCasus())
	require.Equal(t, a.Numerus, a.GetNumerus())
	require.Equal(t, a.Genus, a.GetGenus())
	require.Equal(t, a.Determination, a.GetDetermination())
}
