package identifier

import (
	"strings"
	"unicode"
)

// TextObjectFactoryMaxLength ports TextObjectFactoryBuilder.maxTextLength(10000).
const TextObjectFactoryMaxLength = 10000

// MinorityScriptThreshold ports RemoveMinorityScriptsTextFilter.forThreshold(0.3).
const MinorityScriptThreshold = 0.3

// ApplyTextObjectFactoryFilters ports textObjectFactory.forText(text).toString()
// used in DefaultLanguageIdentifier fallback path (detectLanguageCode).
//
// Java builder order:
//
//	maxTextLength(10000)
//	REMOVE_URL_FILTER
//	RemoveMinorityScriptsTextFilter(0.3)
//	REMOVE_EMAIL_SIGNATURE_FILTER
//	REMOVE_MENTION_FILTER
//	REMOVE_NON_BREAKING_SPACES_FILTER
//
// URL / signature / mention / nbsp match LanguageIdentifier.cleanAndShortenText filters.
func ApplyTextObjectFactoryFilters(text string) string {
	// max length UTF-16
	if javaStringLen(text) > TextObjectFactoryMaxLength {
		text = javaSubstring(text, 0, TextObjectFactoryMaxLength)
	}
	// same filters as cleanAndShorten (without maxLength of identifier)
	text = nbspInvis.ReplaceAllString(text, " ")
	text = mailRE.ReplaceAllString(urlRE.ReplaceAllString(text, " "), " ")
	text = signature.ReplaceAllString(text, "")
	text = mention.ReplaceAllString(text, "")
	text = strings.ReplaceAll(text, "\u00A0", " ")
	text = removeMinorityScripts(text, MinorityScriptThreshold)
	return text
}

// removeMinorityScripts ports optimaize RemoveMinorityScriptsTextFilter.forThreshold.
// Scripts that form less than threshold of letter characters are removed.
func removeMinorityScripts(text string, threshold float64) string {
	if text == "" {
		return text
	}
	// count letters by Unicode script (approx via script ranges / IsLetter groups)
	counts := map[string]int{}
	totalLetters := 0
	runes := []rune(text)
	scripts := make([]string, len(runes))
	for i, r := range runes {
		if !unicode.IsLetter(r) {
			scripts[i] = ""
			continue
		}
		sk := scriptOf(r)
		scripts[i] = sk
		counts[sk]++
		totalLetters++
	}
	if totalLetters == 0 {
		return text
	}
	// scripts below threshold are minority
	minority := map[string]bool{}
	for sk, c := range counts {
		if float64(c)/float64(totalLetters) < threshold {
			minority[sk] = true
		}
	}
	if len(minority) == 0 {
		return text
	}
	var b strings.Builder
	b.Grow(len(text))
	for i, r := range runes {
		if unicode.IsLetter(r) && minority[scripts[i]] {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// scriptOf maps a letter to a coarse script bucket used for minority filtering.
func scriptOf(r rune) string {
	switch {
	case unicode.In(r, unicode.Latin):
		return "Latn"
	case unicode.In(r, unicode.Cyrillic):
		return "Cyrl"
	case unicode.In(r, unicode.Greek):
		return "Grek"
	case unicode.In(r, unicode.Han):
		return "Hani"
	case unicode.In(r, unicode.Hiragana) || unicode.In(r, unicode.Katakana):
		return "Jpan"
	case unicode.In(r, unicode.Hangul):
		return "Hang"
	case unicode.In(r, unicode.Arabic):
		return "Arab"
	case unicode.In(r, unicode.Hebrew):
		return "Hebr"
	case unicode.In(r, unicode.Devanagari):
		return "Deva"
	case unicode.In(r, unicode.Thai):
		return "Thai"
	case unicode.In(r, unicode.Armenian):
		return "Armn"
	case unicode.In(r, unicode.Tamil):
		return "Taml"
	default:
		return "Zyyy"
	}
}
