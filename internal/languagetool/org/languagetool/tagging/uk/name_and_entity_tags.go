package uk

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	// common Ukrainian surname suffixes
	reNameSuffix = regexp.MustCompile(`(?i).+(енко|єнко|ишин|ішин|ський|цький|івна|ович|евич|ів|їв)$`)
	// numbered entities: Т-80, Ан-225, Як-42
	reNumberedEntity = regexp.MustCompile(`^[\p{L}]{1,4}-?\d{1,4}[А-Яа-яA-Za-z]?$`)
	// x-shaped: Т-подібний handled elsewhere; simple letter-dash-letter
	reXShaped = regexp.MustCompile(`(?i)^[\p{L}]-подібн`)
)

// NameSuffixPOS returns a prop name POS for surname-like tokens, or "".
func NameSuffixPOS(token string) string {
	if !reNameSuffix.MatchString(token) {
		return ""
	}
	// require initial capital for prop names
	r, _ := utf8Decode(token)
	if !unicode.IsUpper(r) {
		return ""
	}
	return "noun:anim:m:v_naz:prop:lname"
}

// NumberedEntityPOS tags military/aircraft style designations.
func NumberedEntityPOS(token string) string {
	if reNumberedEntity.MatchString(token) && strings.ContainsAny(token, "0123456789") {
		// must have a letter part
		hasL := false
		for _, r := range token {
			if unicode.IsLetter(r) {
				hasL = true
				break
			}
		}
		if hasL {
			return "noun:inanim:m:v_naz:prop:unanim"
		}
	}
	return ""
}

// AllCapsProperPOS soft-tags ALL-CAPS multi-letter tokens as proper names.
func AllCapsProperPOS(token string) string {
	if len([]rune(token)) < 2 {
		return ""
	}
	for _, r := range token {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return ""
		}
		if !unicode.IsLetter(r) && r != '-' {
			return ""
		}
	}
	// skip pure latin roman numerals (I, V, X, …) already handled as number:latin
	if reLatinNum.MatchString(token) {
		return ""
	}
	return "noun:inanim:m:v_naz:prop"
}

func utf8Decode(s string) (rune, int) {
	for _, r := range s {
		return r, 1
	}
	return 0, 0
}
