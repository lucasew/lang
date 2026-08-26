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
	// Greek
	for _, r := range text {
		if r >= 0x0370 && r <= 0x03FF {
			return "el"
		}
	}
	// Arabic / Persian script
	for _, r := range text {
		if r >= 0x0600 && r <= 0x06FF {
			if strings.ContainsAny(text, "پچژگ") {
				return "fa"
			}
			return "ar"
		}
	}
	// CJK
	for _, r := range text {
		if r >= 0x3040 && r <= 0x30FF {
			return "ja"
		}
		if r >= 0x4E00 && r <= 0x9FFF {
			return "zh"
		}
	}
	// German umlauts / ß
	if strings.ContainsAny(text, "äöüÄÖÜß") {
		return "de"
	}
	// Spanish soft markers
	if strings.ContainsAny(text, "ñ¿¡") ||
		(strings.Contains(lower, " que ") && (strings.Contains(lower, " el ") || strings.Contains(lower, " la "))) {
		return "es"
	}
	// Portuguese soft markers
	if strings.ContainsAny(text, "ãõ") || strings.Contains(lower, " não ") || strings.Contains(lower, " uma ") {
		return "pt"
	}
	// Italian soft markers
	if strings.Contains(lower, " che ") && (strings.Contains(lower, " non ") || strings.Contains(lower, " per ")) {
		return "it"
	}
	// Polish soft markers
	if strings.ContainsAny(text, "ąćęłńóśźżĄĆĘŁŃÓŚŹŻ") {
		return "pl"
	}
	// French accents common words
	if strings.ContainsAny(text, "éèàùç") ||
		(strings.Contains(lower, " le ") && strings.Contains(lower, " est ")) {
		if countLetters(text) > 8 {
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
