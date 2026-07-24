package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// readingWithTagRegex ports AnalyzedTokenReadings.readingWithTagRegex via the core twin
// (Pattern.matcher(pos).matches() full region; first matching reading).
// Empty pattern: Java Pattern.compile("") matches empty tags only — rare; treat as no match
// when tok is nil. Non-nil tok with "" uses core (anchors ^(?:)$).
func readingWithTagRegex(tok *languagetool.AnalyzedTokenReadings, posTagRegex string) *languagetool.AnalyzedToken {
	if tok == nil {
		return nil
	}
	return tok.ReadingWithTagRegex(posTagRegex)
}
