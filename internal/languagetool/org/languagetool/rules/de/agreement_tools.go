package de

import "strings"

// GrammarCategory mirrors AgreementRule.GrammarCategory bits that can be omitted.
type GrammarCategory int

const (
	CatKasus GrammarCategory = iota
	CatNumerus
	CatGenus
)

// AgreementCategoryString builds a compact Kasus/Numerus/Genus key used by
// AgreementTools.getAgreementCategories (POS-free helper for future tagger wiring).
func AgreementCategoryString(casus, numerus, genus string, det string, omit map[GrammarCategory]bool) string {
	var parts []string
	if !omit[CatKasus] && casus != "" {
		parts = append(parts, casus)
	}
	if !omit[CatNumerus] && numerus != "" {
		parts = append(parts, numerus)
	}
	if !omit[CatGenus] && genus != "" {
		parts = append(parts, genus)
	}
	if det != "" {
		parts = append(parts, det)
	}
	return strings.Join(parts, "/")
}

// AgreementTools is the Java-name twin for agreement category string helpers.
type AgreementTools struct{}

func (AgreementTools) CategoryString(casus, numerus, genus, det string, omit map[GrammarCategory]bool) string {
	return AgreementCategoryString(casus, numerus, genus, det, omit)
}
