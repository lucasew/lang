package de

// GermanMultitokenSpeller ports exception logic from
// org.languagetool.rules.de.GermanMultitokenSpeller (without full MultitokenSpeller stack).
type GermanMultitokenSpeller struct{}

// INSTANCE mirrors the Java singleton accessor for call sites.
var GermanMultitokenSpellerInstance = GermanMultitokenSpeller{}

// IsException reports whether original→candidate is a known non-error
// (trailing "s" or "-" stripped).
func (GermanMultitokenSpeller) IsException(original, candidate string) bool {
	if len(original) == 0 || len(candidate) == 0 {
		return false
	}
	if len(original) == len(candidate)+1 && original[:len(original)-1] == candidate {
		last := original[len(original)-1]
		return last == 's' || last == '-'
	}
	return false
}
