package en

import (
	"regexp"
	"strings"
)

// OrdinalSuffixFilter ports org.languagetool.rules.en.OrdinalSuffixFilter.
// Fixes ordinal suggestions like "1nd" → "1st".
type OrdinalSuffixFilter struct{}

func NewOrdinalSuffixFilter() *OrdinalSuffixFilter {
	return &OrdinalSuffixFilter{}
}

var (
	ordinalTeens    = regexp.MustCompile(`.*(11|12|13)$`)
	ordinalNonDigit = regexp.MustCompile(`[^0-9]`)
)

// Fix returns the corrected ordinal string from a broken suggestion (digits extracted).
func (f *OrdinalSuffixFilter) Fix(suggestion string) string {
	ordinal := ordinalNonDigit.ReplaceAllString(suggestion, "")
	if ordinal == "" {
		return suggestion
	}
	if ordinalTeens.MatchString(ordinal) {
		return ordinal + "th"
	}
	switch {
	case strings.HasSuffix(ordinal, "1"):
		return ordinal + "st"
	case strings.HasSuffix(ordinal, "2"):
		return ordinal + "nd"
	case strings.HasSuffix(ordinal, "3"):
		return ordinal + "rd"
	default:
		return ordinal + "th"
	}
}
