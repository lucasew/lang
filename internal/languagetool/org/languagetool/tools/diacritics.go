package tools

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// RemoveDiacritics ports StringTools.removeDiacritics:
// Normalizer.Form.NFD then strip \p{InCombiningDiacriticalMarks} (Mn).
func RemoveDiacritics(str string) string {
	if str == "" {
		return str
	}
	s := norm.NFD.String(str)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// HasDiacritics ports StringTools.hasDiacritics.
func HasDiacritics(str string) bool {
	return str != RemoveDiacritics(str)
}

// EqualsIgnoreCaseAndDiacritics ports StringTools.equalsIgnoreCaseAndDiacritics.
func EqualsIgnoreCaseAndDiacritics(s1, s2 string) bool {
	// Java: null equality; Go uses empty-string only for present strings.
	return strings.EqualFold(RemoveDiacritics(s1), RemoveDiacritics(s2))
}
