package de

import "strings"

// IsVowel ports GermanTools.isVowel.
func IsVowel(c rune) bool {
	return strings.ContainsRune("aeiouyAEIOUYäöüÄÖÜ", c)
}
