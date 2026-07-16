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

// HasReadingOfType reports whether any reading matches the German POS type
// (prefix-based stand-in for AnalyzedGermanToken.getType()).
func HasReadingOfType(readings *languagetool.AnalyzedTokenReadings, typ POSType) bool {
	if readings == nil {
		return false
	}
	for _, tok := range readings.GetReadings() {
		pt := tok.GetPOSTag()
		if pt == nil {
			continue
		}
		tag := *pt
		if tag == languagetool.SentenceEndTagName || tag == languagetool.ParagraphEndTagName {
			return false
		}
		if posTypeFromTag(tag) == typ {
			return true
		}
	}
	return false
}

func posTypeFromTag(tag string) POSType {
	switch {
	case strings.HasPrefix(tag, "SUB"):
		return POSNomen
	case strings.HasPrefix(tag, "VER"):
		return POSVerb
	case strings.HasPrefix(tag, "ADJ") || strings.HasPrefix(tag, "PA"):
		return POSAdjective
	case strings.HasPrefix(tag, "ART"):
		return POSDeterminer
	case strings.HasPrefix(tag, "PRO"):
		return POSPronoun
	case strings.HasPrefix(tag, "ADV"):
		return POSAdverb
	case strings.HasPrefix(tag, "KON"):
		return POSConjunction
	case strings.HasPrefix(tag, "PRP") || strings.HasPrefix(tag, "APPR"):
		return POSPreposition
	case strings.HasPrefix(tag, "ZAL") || strings.HasPrefix(tag, "CARD"):
		return POSNumeral
	case strings.HasPrefix(tag, "ITJ"):
		return POSInterjection
	case strings.HasPrefix(tag, "PTK"):
		return POSParticle
	default:
		return POSOther
	}
}
