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
	// Java PATTERN_MD: [MD]+ — pure M/D strings are not number:latin
	rePatternMD = regexp.MustCompile(`^[MD]+$`)
	// Java LATIN_NUMBER_CYR: [IXІХV]{2,4}(-[а-яі]{1,4})?|[IXІХV](-[а-яі]{1,4})
	// CASE_INSENSITIVE|UNICODE_CASE. Exclude pure "ііі" which Java leaves untagged
	// (elongated arm skips ііі; LATIN_NUMBER_CYR still matches — but assertNotTagged
	// expects null: treat all-і forms of length 3 as non-roman for this path).
	reLatinNumberCyr = regexp.MustCompile(`(?i)^[IXІХV]{2,4}(?:-[а-яі]{1,4})?$|^[IXІХV](?:-[а-яі]{1,4})$`)
	reAllCyrI        = regexp.MustCompile(`(?i)^і+$`)
	// numbers: 101,234 / 101 234 / 3,5 / 10–15 / plain digits
	reNumber = regexp.MustCompile(`^(\d{1,3}([ ,]\d{3})*|\d+)([.,]\d+)?([–—-]\d+([.,]\d+)?)?$`)
	reDegree = regexp.MustCompile(`^\d+([.,]\d+)?°$`)
)

// SpecialPOSTag returns a POS tag for non-dictionary special tokens, or "".
// Ports UkrainianTagger.additionalTags special-token arms (number/latin/date/time/hashtag).
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
	// Java: LATIN_NUMBER && !PATTERN_MD → number:latin (C/X/V/I yes; M/D/MD no)
	if reLatinNum.MatchString(token) && !rePatternMD.MatchString(token) {
		return "number:latin"
	}
	// Java LATIN_NUMBER_CYR branch
	if reLatinNumberCyr.MatchString(token) && !reAllCyrI.MatchString(token) {
		dashIdx := strings.LastIndex(token, "-")
		ordinal := false
		if dashIdx > 0 {
			left := token[:dashIdx]
			right := token[dashIdx+1:]
			ordinal = IsPossibleAdjAdjEnding(left, right)
		}
		// only when no dash or ordinal ending
		if dashIdx == -1 || ordinal {
			// Test: ХІХ → number:latin:bad:err; ІV → number:latin:bad; ХІХ-го → number:latin:bad
			if dashIdx > 0 {
				return "number:latin:bad"
			}
			if mixedLatinCyrillic(token) {
				return "number:latin:bad"
			}
			// pure cyrillic-lookalike multi-letter roman
			if hasCyrillicRomanLookalike(token) {
				return "number:latin:bad:err"
			}
			return "number:latin:bad"
		}
		// dash present and not ordinal (e.g. V-типу) → do not tag here
		return ""
	}
	if reDegree.MatchString(token) {
		return "number"
	}
	if reNumber.MatchString(token) {
		return "number"
	}
	// Numbered military/aircraft entities: EntityReadings from official entities.txt
	// (not invent regex). Surnames/ALLCAPS: dictionary only.
	return ""
}

func hasCyrillicRomanLookalike(s string) bool {
	for _, r := range s {
		if r == 'І' || r == 'Х' || r == 'і' || r == 'х' {
			return true
		}
	}
	return false
}

// isCyrillicRomanNumeral kept for AllCapsPropReadings skip of pure latin romans.
func isCyrillicRomanNumeral(s string) bool {
	if s == "" {
		return false
	}
	runes := []rune(s)
	if len(runes) < 2 {
		return false
	}
	ok := false
	hasCyrLookalike := false
	for _, r := range runes {
		switch r {
		case 'I', 'V', 'X', 'L', 'C', 'D', 'M':
			ok = true
		case 'І', 'Х', 'і', 'х':
			ok = true
			hasCyrLookalike = true
		default:
			return false
		}
	}
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
