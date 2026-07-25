package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
	"github.com/stretchr/testify/require"
)

func atrAgreement(token, pos, lemma string) *languagetool.AnalyzedTokenReadings {
	var p, l *string
	if pos != "" {
		p = strPtr(pos)
	}
	if lemma != "" {
		l = strPtr(lemma)
	}
	return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(token, p, l), 0)
}

// Twin: AgreementTools.makeString uses GermanToken.*.toString() → German names + "definit"/"indefinit".
func TestAgreementTools_MakeString(t *testing.T) {
	s := makeString(detag.KasusNominativ, detag.NumerusSingular, detag.GenusMaskulinum, detag.DeterminationDefinite, nil)
	require.Equal(t, "Nominativ/Singular/Maskulinum/definit", s)

	// omit GENUS → Java !omit.contains(GENUS) false for genus
	s2 := makeString(detag.KasusNominativ, detag.NumerusSingular, detag.GenusMaskulinum, detag.DeterminationDefinite,
		map[GrammarCategory]bool{CatGenus: true})
	require.Equal(t, "Nominativ/Singular/definit", s2)

	s3 := makeString(detag.KasusNominativ, detag.NumerusSingular, detag.GenusMaskulinum, detag.DeterminationDefinite,
		map[GrammarCategory]bool{CatKasus: true, CatNumerus: true})
	require.Equal(t, "Maskulinum/definit", s3)

	// null casus omitted (empty string twin of null)
	s4 := makeString("", detag.NumerusSingular, detag.GenusFemininum, detag.DeterminationIndefinite, nil)
	require.Equal(t, "Singular/Femininum/indefinit", s4)
}

// Java getAgreementCategories(ART:DEF:NOM:SIN:FEM) → single key with definit (determination non-null).
func TestAgreementTools_GetAgreementCategories_DefiniteArticle(t *testing.T) {
	det := atrAgreement("die", "ART:DEF:NOM:SIN:FEM", "die")
	cats := GetAgreementCategories(det, nil, false)
	require.Equal(t, map[string]struct{}{
		"Nominativ/Singular/Femininum/definit": {},
	}, cats)
}

// Java: noun determination null → both DEF and IND.
func TestAgreementTools_GetAgreementCategories_NounBothDeterminations(t *testing.T) {
	noun := atrAgreement("Haus", "SUB:NOM:SIN:NEU", "Haus")
	cats := GetAgreementCategories(noun, nil, false)
	require.Equal(t, map[string]struct{}{
		"Nominativ/Singular/Neutrum/definit":   {},
		"Nominativ/Singular/Neutrum/indefinit": {},
	}, cats)
}

// Java: genus ALG + determination null → expand 3 genders × 2 determinations.
func TestAgreementTools_GetAgreementCategories_ALGExpand(t *testing.T) {
	// PRO:IND:NOM:PLU:ALG — IND sets determination non-null, so only 3 genders with indefinit.
	proInd := atrAgreement("alle", "PRO:IND:NOM:PLU:ALG", "alle")
	cats := GetAgreementCategories(proInd, nil, false)
	require.Equal(t, map[string]struct{}{
		"Nominativ/Plural/Maskulinum/indefinit": {},
		"Nominativ/Plural/Femininum/indefinit":  {},
		"Nominativ/Plural/Neutrum/indefinit":    {},
	}, cats)

	// ALG without DEF/IND part — rare tag shape; use SUB with ALG (Morphy sometimes).
	// If determination null: 6 strings.
	algNoun := atrAgreement("X", "SUB:NOM:SIN:ALG", "X")
	cats2 := GetAgreementCategories(algNoun, nil, false)
	require.Equal(t, map[string]struct{}{
		"Nominativ/Singular/Maskulinum/definit":   {},
		"Nominativ/Singular/Maskulinum/indefinit": {},
		"Nominativ/Singular/Femininum/definit":    {},
		"Nominativ/Singular/Femininum/indefinit":  {},
		"Nominativ/Singular/Neutrum/definit":      {},
		"Nominativ/Singular/Neutrum/indefinit":    {},
	}, cats2)
}

// Java: ALG + :STV → do not expand genders (stellvertretend).
func TestAgreementTools_GetAgreementCategories_ALG_STV_NoExpand(t *testing.T) {
	// ART:DEF:NOM:SIN:ALG:STV — STV blocks ALG expansion; determination DEF → one key with Allgemein.
	tok := atrAgreement("d.", "ART:DEF:NOM:SIN:ALG:STV", "der")
	cats := GetAgreementCategories(tok, nil, false)
	require.Equal(t, map[string]struct{}{
		"Nominativ/Singular/Allgemein/definit": {},
	}, cats)
}

// Java: "jed"/"manch" lemma → both determinations even when tag has DEF/IND.
func TestAgreementTools_GetAgreementCategories_JedManch(t *testing.T) {
	jed := atrAgreement("jeder", "PRO:IND:NOM:SIN:MAS", "jed")
	cats := GetAgreementCategories(jed, nil, false)
	require.Equal(t, map[string]struct{}{
		"Nominativ/Singular/Maskulinum/definit":   {},
		"Nominativ/Singular/Maskulinum/indefinit": {},
	}, cats)

	manch := atrAgreement("manche", "PRO:IND:NOM:PLU:FEM", "manch")
	cats2 := GetAgreementCategories(manch, nil, false)
	require.Equal(t, map[string]struct{}{
		"Nominativ/Plural/Femininum/definit":   {},
		"Nominativ/Plural/Femininum/indefinit": {},
	}, cats2)
}

// Java skipSol: readings ending :SOL skipped when skipSol true.
func TestAgreementTools_GetAgreementCategories_SkipSol(t *testing.T) {
	pSol := "ADJ:NOM:SIN:NEU:GRU:SOL"
	pInd := "ADJ:NOM:SIN:NEU:GRU:IND"
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("rotes", &pSol, strPtr("rot")), 0)
	tok.AddReading(languagetool.NewAnalyzedToken("rotes", &pInd, strPtr("rot")), "")

	withSkip := GetAgreementCategories(tok, nil, true)
	// only IND reading (non-SOL)
	require.Equal(t, map[string]struct{}{
		"Nominativ/Singular/Neutrum/indefinit": {},
	}, withSkip)

	noSkip := GetAgreementCategories(tok, nil, false)
	// SOL has no DEF/IND → both determinations; IND reading adds indefinit again
	require.Contains(t, noSkip, "Nominativ/Singular/Neutrum/definit")
	require.Contains(t, noSkip, "Nominativ/Singular/Neutrum/indefinit")
	require.Len(t, noSkip, 2)
}

// Java getAgreementSOLCategories: only :SOL readings; both DEF and IND.
func TestAgreementTools_GetAgreementSOLCategories(t *testing.T) {
	pSol := "ADJ:NOM:SIN:NEU:GRU:SOL"
	pInd := "ADJ:NOM:SIN:NEU:GRU:IND"
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("rotes", &pSol, strPtr("rot")), 0)
	tok.AddReading(languagetool.NewAnalyzedToken("rotes", &pInd, strPtr("rot")), "")
	cats := GetAgreementSOLCategories(tok, nil)
	require.Equal(t, map[string]struct{}{
		"Nominativ/Singular/Neutrum/definit":   {},
		"Nominativ/Singular/Neutrum/indefinit": {},
	}, cats)

	tok2 := atrAgreement("rote", pInd, "rot")
	require.Empty(t, GetAgreementSOLCategories(tok2, nil))
}

// Java possessiveSpecialCase: PRO:POS + lemma ich/sich with ALG → do not expand (else branch).
func TestAgreementTools_PossessiveSpecialCase_IchSich(t *testing.T) {
	// PRO:POS:...:ALG with lemma "ich" — ALG expand skipped due to possessiveSpecialCase.
	// Tag must parse casus/numerus/genus; ALG + no DEF/IND → determination null → both dets, genus stays ALG.
	tok := atrAgreement("meiner", "PRO:POS:GEN:SIN:ALG", "ich")
	require.True(t, tok.HasPosTagStartingWith("PRO:POS"))
	cats := GetAgreementCategories(tok, nil, false)
	require.Equal(t, map[string]struct{}{
		"Genitiv/Singular/Allgemein/definit":   {},
		"Genitiv/Singular/Allgemein/indefinit": {},
	}, cats)

	// lemma "sich" same
	tok2 := atrAgreement("seiner", "PRO:POS:GEN:SIN:ALG", "sich")
	cats2 := GetAgreementCategories(tok2, nil, false)
	require.Equal(t, map[string]struct{}{
		"Genitiv/Singular/Allgemein/definit":   {},
		"Genitiv/Singular/Allgemein/indefinit": {},
	}, cats2)

	// other lemma with PRO:POS ALG → still expands (not special case)
	tok3 := atrAgreement("meiner", "PRO:POS:GEN:SIN:ALG", "mein")
	cats3 := GetAgreementCategories(tok3, nil, false)
	require.Contains(t, cats3, "Genitiv/Singular/Maskulinum/definit")
	require.Contains(t, cats3, "Genitiv/Singular/Femininum/definit")
	require.Contains(t, cats3, "Genitiv/Singular/Neutrum/definit")
	require.NotContains(t, cats3, "Genitiv/Singular/Allgemein/definit")
}

// omit Kasus: keys without Nominativ etc.
func TestAgreementTools_OmitCategories(t *testing.T) {
	det := atrAgreement("die", "ART:DEF:NOM:SIN:FEM", "die")
	omit := map[GrammarCategory]bool{CatKasus: true}
	cats := GetAgreementCategories(det, omit, false)
	require.Equal(t, map[string]struct{}{
		"Singular/Femininum/definit": {},
	}, cats)
}

// Readings without casus/numerus/genus contribute nothing.
func TestAgreementTools_SkipEmptyMorph(t *testing.T) {
	tok := atrAgreement("und", "KON", "und")
	require.Empty(t, GetAgreementCategories(tok, nil, false))
	require.Empty(t, GetAgreementSOLCategories(tok, nil))
}

// die(FEM) vs Haus(NEU) — no common category string (Java retainAll → empty).
func TestAgreementTools_DetNounGenderMismatchKeys(t *testing.T) {
	det := atrAgreement("die", "ART:DEF:NOM:SIN:FEM", "die")
	noun := atrAgreement("Haus", "SUB:NOM:SIN:NEU", "Haus")
	cats := GetAgreementCategories(det, nil, false)
	ncats := GetAgreementCategories(noun, nil, false)
	require.Contains(t, cats, "Nominativ/Singular/Femininum/definit")
	require.Contains(t, ncats, "Nominativ/Singular/Neutrum/definit")
	for k := range cats {
		_, ok := ncats[k]
		require.False(t, ok, "unexpected shared key %q", k)
	}
}
