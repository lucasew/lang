package commandline

import (
	"strings"
	"unicode"
)

// DetectLanguageFunc detects a language code from text.
type DetectLanguageFunc func(text string) string

// DetectLanguageHeuristic is a tiny script/keyword heuristic for green tests
// (full FastText/SimpleLanguageIdentifier wired by hooks in production).
func DetectLanguageHeuristic(text string) string {
	lower := strings.ToLower(text)
	// Cyrillic → uk/ru soft: prefer uk if Ukrainian letters present
	hasCyr := false
	hasUkr := false
	for _, r := range text {
		if r >= 0x0400 && r <= 0x04FF {
			hasCyr = true
		}
		if strings.ContainsRune("іїєґІЇЄҐ", r) {
			hasUkr = true
		}
	}
	if hasUkr {
		return "uk"
	}
	if hasCyr {
		return "ru"
	}
	// German umlauts / ß
	if strings.ContainsAny(text, "äöüÄÖÜß") {
		return "de"
	}
	// French accents common words
	if strings.Contains(lower, " le ") || strings.Contains(lower, " la ") ||
		strings.Contains(lower, "est ") || strings.ContainsAny(text, "éèàùç") {
		// weak FR signal
		if countLetters(text) > 10 {
			return "fr"
		}
	}
	// default English
	_ = unicode.IsLetter
	return "en"
}

func countLetters(s string) int {
	n := 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			n++
		}
	}
	return n
}

// ResolveLanguage returns opts.Language or autodetection when AutoDetect is set.
func ResolveLanguage(text string, opts *CommandLineOptions, detect DetectLanguageFunc) string {
	if opts != nil && opts.Language != "" && !opts.AutoDetect {
		return opts.Language
	}
	if opts != nil && opts.AutoDetect {
		fn := detect
		if fn == nil {
			fn = DetectLanguageHeuristic
		}
		return fn(text)
	}
	if opts != nil && opts.Language != "" {
		return opts.Language
	}
	return "en"
}
