package uk

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	reHashtag   = regexp.MustCompile(`^#[\p{L}\p{N}_]+$`)
	reDate      = regexp.MustCompile(`^\d{1,2}\.\d{1,2}\.\d{2,4}$`)
	reTimeColon = regexp.MustCompile(`^\d{1,2}:\d{2}(:\d{2})?$`)
	reTimeDot   = regexp.MustCompile(`^\d{1,2}\.\d{2}$`)
	reLatinNum  = regexp.MustCompile(`^[IVXLCDM]+$`)
	// numbers: 101,234 / 101 234 / 3,5 / 10–15 / plain digits
	reNumber = regexp.MustCompile(`^(\d{1,3}([ ,]\d{3})*|\d+)([.,]\d+)?([–—-]\d+([.,]\d+)?)?$`)
	reDegree = regexp.MustCompile(`^\d+([.,]\d+)?°$`)
)

// SpecialPOSTag returns a POS tag for non-dictionary special tokens, or "".
func SpecialPOSTag(token string) string {
	if token == "" {
		return ""
	}
	if reHashtag.MatchString(token) {
		return "hashtag"
	}
	if reDate.MatchString(token) {
		return "date"
	}
	if reTimeColon.MatchString(token) {
		return "time"
	}
	// "15.33" as time when looks like HH.MM (not date - date needs 3 parts)
	if reTimeDot.MatchString(token) {
		parts := strings.Split(token, ".")
		if len(parts) == 2 && len(parts[0]) <= 2 && len(parts[1]) == 2 {
			return "time"
		}
	}
	if reLatinNum.MatchString(token) {
		// single D/C are often not roman in isolation (Java: D → null, X → number:latin)
		if len(token) == 1 && (token == "D" || token == "C" || token == "L" || token == "M") {
			return ""
		}
		return "number:latin"
	}
	// e.g. ХІХ-го
	if i := strings.Index(token, "-"); i > 0 {
		base := token[:i]
		if isCyrillicRomanNumeral(base) || reLatinNum.MatchString(base) {
			return "number:latin:bad"
		}
	}
	// Cyrillic lookalike romans: ХІХ, ІV
	if isCyrillicRomanNumeral(token) {
		if mixedLatinCyrillic(token) {
			return "number:latin:bad"
		}
		return "number:latin:bad:err"
	}
	if reDegree.MatchString(token) {
		return "number"
	}
	if token == "C" {
		return "number:latin"
	}
	if reNumber.MatchString(token) {
		return "number"
	}
	// Numbered military/aircraft entities: EntityReadings from official entities.txt
	// (not invent regex). Surnames/ALLCAPS: dictionary only.
	return ""
}

func isCyrillicRomanNumeral(s string) bool {
	if s == "" {
		return false
	}
	// only roman-like glyphs (latin + cyrillic lookalikes І/Х), not arbitrary cyrillic words
	ok := false
	hasCyrLookalike := false
	for _, r := range s {
		switch r {
		case 'I', 'V', 'X', 'L', 'C', 'D', 'M':
			ok = true
		case 'І', 'Х', 'і', 'х': // cyrillic I/X lookalikes
			ok = true
			hasCyrLookalike = true
		default:
			return false
		}
	}
	// pure latin handled by reLatinNum; here we want mixed or cyrillic-only romans
	return ok && (hasCyrLookalike || mixedLatinCyrillic(s)) && !reLatinNum.MatchString(s)
}

func mixedLatinCyrillic(s string) bool {
	hasLat, hasCyr := false, false
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			hasLat = true
		}
		if unicode.Is(unicode.Cyrillic, r) {
			hasCyr = true
		}
	}
	return hasLat && hasCyr
}
