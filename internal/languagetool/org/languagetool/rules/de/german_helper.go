package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// POSType mirrors GermanToken.POSType for helper checks.
type POSType int

const (
	POSOther POSType = iota
	POSNomen
	POSVerb
	POSAdjective
	POSDeterminer
	POSPronoun
	POSPartizip
	POSProperNoun
	// Remaining constants kept for callers that may reference them; not set by
	// AnalyzedGermanToken parsing (Java leaves type null for ADV/KON/…).
	POSParticle
	POSAdverb
	POSConjunction
	POSNumeral
	POSInterjection
	POSPreposition
)

// GermanHelper ports org.languagetool.rules.de.GermanHelper POS-tag utilities.
func GetNounCase(posTag string) string   { return getIndexOrEmpty(posTag, 1) }
func GetNounNumber(posTag string) string { return getIndexOrEmpty(posTag, 2) }
func GetNounGender(posTag string) string { return getIndexOrEmpty(posTag, 3) }

func GetDeterminerDefiniteness(posTag string) string { return getIndexOrEmpty(posTag, 1) }
func GetDeterminerCase(posTag string) string         { return getIndexOrEmpty(posTag, 2) }
func GetDeterminerNumber(posTag string) string       { return getIndexOrEmpty(posTag, 3) }
func GetDeterminerGender(posTag string) string       { return getIndexOrEmpty(posTag, 4) }

// GetComparison returns GRU, KOM, or SUP (or empty).
func GetComparison(posTag string) string {
	cmp := getIndexOrEmpty(posTag, 4)
	if cmp != "GRU" && cmp != "KOM" && cmp != "SUP" {
		cmp = getIndexOrEmpty(posTag, 2)
	}
	return cmp
}

func getIndexOrEmpty(posTag string, idx int) string {
	if posTag == "" {
		return ""
	}
	parts := strings.Split(posTag, ":")
	if len(parts) > idx {
		return parts[idx]
	}
	return ""
}

// HasReadingOfType ports GermanHelper.hasReadingOfType via AnalyzedGermanToken.getType().
func HasReadingOfType(readings *languagetool.AnalyzedTokenReadings, typ POSType) bool {
	if readings == nil {
		return false
	}
	for _, tok := range readings.GetReadings() {
		if tok == nil {
			continue
		}
		pt := tok.GetPOSTag()
		if pt != nil &&
			(*pt == languagetool.SentenceEndTagName || *pt == languagetool.ParagraphEndTagName) {
			return false
		}
		if posTypeFromAnalyzedGermanToken(pt) == typ {
			return true
		}
	}
	return false
}

// posTypeFromAnalyzedGermanToken ports AnalyzedGermanToken constructor type resolution.
// Java: only tags with ≥3 colon parts get a type; EIG/SUB/PA/VER/ADJ/PRO/ART map as below.
func posTypeFromAnalyzedGermanToken(posTag *string) POSType {
	if posTag == nil || *posTag == "" {
		return POSOther
	}
	parts := strings.Split(*posTag, ":")
	if len(parts) < 3 {
		return POSOther
	}
	// Java AnalyzedGermanToken: tempType starts null; PA1/PA2 always assign; others only if null.
	var temp *POSType
	for _, part := range parts {
		switch part {
		case "EIG":
			t := POSProperNoun
			temp = &t
		case "SUB":
			if temp == nil {
				t := POSNomen
				temp = &t
			}
		case "PA1", "PA2":
			t := POSPartizip
			temp = &t
		case "VER":
			if temp == nil {
				t := POSVerb
				temp = &t
			}
		case "ADJ":
			if temp == nil {
				t := POSAdjective
				temp = &t
			}
		case "PRO":
			if temp == nil {
				t := POSPronoun
				temp = &t
			}
		case "ART":
			if temp == nil {
				t := POSDeterminer
				temp = &t
			}
		}
	}
	if temp == nil {
		return POSOther
	}
	return *temp
}
