package de

import (
	"regexp"
	"strings"
)

// Gender-neutral compound validation for ignorePotentiallyMisspelledWord
// (GermanSpellerRule.isValidGenderNeutralWord + GENDER2_STAR2 gate).

var (
	// GENDER2_STAR2: word contains Binnen-I or *:_in or /-in marker.
	// Java: .*((?<=(\w))In|[\*:_]in|/-in).* — RE2 has no lookbehind.
	reGender2Star2 = regexp.MustCompile(`(?:[0-9A-Za-zÄÖÜäöüß]In|[\*:_]in|/-in)`)
	// POTENTIAL_BINNEN_I on a part window
	rePotentialBinnenI = regexp.MustCompile(`[0-9A-Za-zÄÖÜäöüß]In`)
	reGenderNeutralSin = regexp.MustCompile(`[\*:_/]in$`)
	reGenderNeutralPlu = regexp.MustCompile(`[\*:_/]in.`)
	reGenderNeutralSlashHyphen = regexp.MustCompile(`/-in.`)
	// replaceFirst helpers
	reBinnenIReplace   = regexp.MustCompile(`([0-9A-Za-zÄÖÜäöüß])In`)
	reSpecialInReplace = regexp.MustCompile(`[\*:_/]in`)
	reSlashHyphenIn    = regexp.MustCompile(`/-in`)
)

// hasGender2Star2 ports GENDER2_STAR2.matcher(word).matches() (contains marker).
func hasGender2Star2(word string) bool {
	return reGender2Star2.MatchString(word)
}

// replaceFirstBinnenI ports replaceFirst("((?<=(\\w))In)", "in") once.
func replaceFirstBinnenI(s string) string {
	loc := reBinnenIReplace.FindStringSubmatchIndex(s)
	if loc == nil {
		return s
	}
	letter := s[loc[2]:loc[3]]
	return s[:loc[0]] + letter + "in" + s[loc[1]:]
}

func replaceFirstSpecialIn(s string) string {
	loc := reSpecialInReplace.FindStringIndex(s)
	if loc == nil {
		return s
	}
	return s[:loc[0]] + "in" + s[loc[1]:]
}

func replaceFirstSlashHyphenIn(s string) string {
	loc := reSlashHyphenIn.FindStringIndex(s)
	if loc == nil {
		return s
	}
	return s[:loc[0]] + "in" + s[loc[1]:]
}

// isValidGenderNeutralWord ports GermanSpellerRule.isValidGenderNeutralWord.
// parts are tokenizer parts of the (gender-normalized) form; word is the original
// surface before gender normalization (wordNoDotOrg), as in Java.
func (r *GermanSpellerRule) isValidGenderNeutralWord(parts []string, word string) bool {
	if r == nil || word == "" || len(parts) == 0 {
		return false
	}
	// Java uses String indices (UTF-16); BMP-only DE gender forms → rune index is OK for umlauts.
	runes := []rune(word)
	start := 0
	for _, part := range parts {
		partRunes := []rune(part)
		end := start + len(partRunes)
		if end > len(runes) {
			return false
		}
		toCheck := string(runes[start:end])
		if strings.HasPrefix(toCheck, "I") && start > 0 {
			// e.g. AktienIndex
			return false
		}
		if rePotentialBinnenI.MatchString(toCheck) {
			norm := replaceFirstBinnenI(toCheck)
			if r.IsMisspelled(norm) || (!strings.HasSuffix(toCheck, "In") && !strings.HasSuffix(toCheck, "Innen")) {
				return false
			}
		}
		if reGenderNeutralSin.MatchString(toCheck) {
			if r.IsMisspelled(replaceFirstSpecialIn(toCheck)) {
				return false
			}
			end++
		}
		if reGenderNeutralPlu.MatchString(toCheck) {
			if end < len(runes) {
				end++
				toCheck = string(runes[start:end])
			}
			if r.IsMisspelled(replaceFirstSpecialIn(toCheck)) ||
				(!strings.HasSuffix(toCheck, "in") && !strings.HasSuffix(toCheck, "innen")) {
				return false
			}
		}
		if reGenderNeutralSlashHyphen.MatchString(toCheck) {
			if end+1 < len(runes) {
				end += 2
				toCheck = string(runes[start:end])
			}
			if r.IsMisspelled(replaceFirstSlashHyphenIn(toCheck)) ||
				(!strings.HasSuffix(toCheck, "in") && !strings.HasSuffix(toCheck, "innen")) {
				return false
			}
		}
		start = end
	}
	return true
}
