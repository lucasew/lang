package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
)

// GrammarCategory mirrors AgreementRule.GrammarCategory bits that can be omitted.
type GrammarCategory int

const (
	CatKasus GrammarCategory = iota
	CatNumerus
	CatGenus
)

// AgreementCategoryString builds a compact Kasus/Numerus/Genus key used by
// AgreementTools.getAgreementCategories.
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

// GetAgreementCategories ports AgreementTools.getAgreementCategories for tagged tokens.
// omit may skip Kasus/Numerus/Genus; skipSol skips :SOL adjective readings.
func GetAgreementCategories(aToken *languagetool.AnalyzedTokenReadings, omit map[GrammarCategory]bool, skipSol bool) map[string]struct{} {
	set := map[string]struct{}{}
	if aToken == nil {
		return set
	}
	if omit == nil {
		omit = map[GrammarCategory]bool{}
	}
	for _, tmpReading := range aToken.GetReadings() {
		if tmpReading == nil || tmpReading.GetPOSTag() == nil {
			continue
		}
		pos := *tmpReading.GetPOSTag()
		if skipSol && strings.HasSuffix(pos, ":SOL") {
			continue
		}
		reading := detag.ParseGermanPOS(pos)
		if reading.Kasus == "" && reading.Numerus == "" && reading.Genus == "" {
			continue
		}
		// ALG expands to all genders (Java Genus.ALLGEMEIN)
		if reading.Genus == detag.GenusALG && !strings.HasSuffix(pos, ":STV") &&
			!possessiveSpecialCase(aToken, tmpReading) {
			gens := []detag.Genus{detag.GenusMas, detag.GenusFem, detag.GenusNeu}
			if reading.Determination == "" {
				for _, g := range gens {
					set[makeAgreementString(reading.Kasus, reading.Numerus, g, detag.DetDefinite, omit)] = struct{}{}
					set[makeAgreementString(reading.Kasus, reading.Numerus, g, detag.DetIndefinite, omit)] = struct{}{}
				}
			} else {
				for _, g := range gens {
					set[makeAgreementString(reading.Kasus, reading.Numerus, g, reading.Determination, omit)] = struct{}{}
				}
			}
			continue
		}
		// jed/manch: both DEF and IND
		lemma := ""
		if tmpReading.GetLemma() != nil {
			lemma = *tmpReading.GetLemma()
		}
		if reading.Determination == "" || lemma == "jed" || lemma == "manch" {
			set[makeAgreementString(reading.Kasus, reading.Numerus, reading.Genus, detag.DetDefinite, omit)] = struct{}{}
			set[makeAgreementString(reading.Kasus, reading.Numerus, reading.Genus, detag.DetIndefinite, omit)] = struct{}{}
		} else {
			set[makeAgreementString(reading.Kasus, reading.Numerus, reading.Genus, reading.Determination, omit)] = struct{}{}
		}
	}
	return set
}

func makeAgreementString(casus detag.Kasus, num detag.Numerus, gen detag.Genus, det detag.Determination, omit map[GrammarCategory]bool) string {
	var l []string
	if casus != "" && !omit[CatKasus] {
		l = append(l, string(casus))
	}
	if num != "" && !omit[CatNumerus] {
		l = append(l, string(num))
	}
	if gen != "" && !omit[CatGenus] {
		l = append(l, string(gen))
	}
	if det != "" {
		l = append(l, string(det))
	}
	return strings.Join(l, "/")
}

func possessiveSpecialCase(aToken *languagetool.AnalyzedTokenReadings, tmpReading *languagetool.AnalyzedToken) bool {
	if aToken == nil || !aToken.HasPosTagStartingWith("PRO:POS") {
		return false
	}
	if tmpReading == nil || tmpReading.GetLemma() == nil {
		return false
	}
	lem := *tmpReading.GetLemma()
	return lem == "ich" || lem == "sich"
}

// GetAgreementSOLCategories ports AgreementTools.getAgreementSOLCategories:
// only readings whose POS ends with :SOL (alleinstehend).
func GetAgreementSOLCategories(aToken *languagetool.AnalyzedTokenReadings, omit map[GrammarCategory]bool) map[string]struct{} {
	set := map[string]struct{}{}
	if aToken == nil {
		return set
	}
	if omit == nil {
		omit = map[GrammarCategory]bool{}
	}
	for _, tmpReading := range aToken.GetReadings() {
		if tmpReading == nil || tmpReading.GetPOSTag() == nil {
			continue
		}
		pos := *tmpReading.GetPOSTag()
		if !strings.HasSuffix(pos, ":SOL") {
			continue
		}
		reading := detag.ParseGermanPOS(pos)
		if reading.Kasus == "" && reading.Numerus == "" && reading.Genus == "" {
			continue
		}
		set[makeAgreementString(reading.Kasus, reading.Numerus, reading.Genus, detag.DetDefinite, omit)] = struct{}{}
		set[makeAgreementString(reading.Kasus, reading.Numerus, reading.Genus, detag.DetIndefinite, omit)] = struct{}{}
	}
	return set
}

// CategoriesIntersect reports whether two category sets share any key.
// Empty side → treated as non-mismatch for call sites that fail-closed on missing tags
// (Java checkDetNoun uses retainAll on empty differently; STV is handled separately).
func CategoriesIntersect(a, b map[string]struct{}) bool {
	if len(a) == 0 || len(b) == 0 {
		return true
	}
	for k := range a {
		if _, ok := b[k]; ok {
			return true
		}
	}
	return false
}

// AgreementTools is the Java-name twin for agreement category string helpers.
type AgreementTools struct{}

func (AgreementTools) CategoryString(casus, numerus, genus, det string, omit map[GrammarCategory]bool) string {
	return AgreementCategoryString(casus, numerus, genus, det, omit)
}

func (AgreementTools) GetCategories(aToken *languagetool.AnalyzedTokenReadings, omit map[GrammarCategory]bool, skipSol bool) map[string]struct{} {
	return GetAgreementCategories(aToken, omit, skipSol)
}

func (AgreementTools) GetSOLCategories(aToken *languagetool.AnalyzedTokenReadings, omit map[GrammarCategory]bool) map[string]struct{} {
	return GetAgreementSOLCategories(aToken, omit)
}
