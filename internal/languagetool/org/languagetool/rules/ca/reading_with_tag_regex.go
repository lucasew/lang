package ca

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// readingWithTagRegex ports AnalyzedTokenReadings.readingWithTagRegex (first matching reading).
func readingWithTagRegex(tok *languagetool.AnalyzedTokenReadings, posTagRegex string) *languagetool.AnalyzedToken {
	if tok == nil || posTagRegex == "" {
		return nil
	}
	re, err := regexp.Compile(posTagRegex)
	if err != nil {
		return nil
	}
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		// Java String.matches → full match
		tag := *r.GetPOSTag()
		loc := re.FindStringIndex(tag)
		if loc != nil && loc[0] == 0 && loc[1] == len(tag) {
			return r
		}
	}
	return nil
}
