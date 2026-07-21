package server

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/commandline"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
)

// DetectLanguageOfString ports TextChecker.detectLanguageOfString variant resolution
// after languageIdentifier.detectLanguage (injectable via detect).
//
// Java twin (TextChecker.java):
//  1. detected = identifier.detectLanguage(cleanText, …); if null → parseLanguage(fallback or "en")
//  2. if preferredVariants non-empty: for each, require dash; if preferred.split("-")[0].equals(lang.getShortCode()) → parseLanguage(preferred)
//  3. else if preferred empty: lang = lang.getDefaultLanguageVariant() when non-null
//
// detect may be nil → commandline.DetectLanguageHeuristic.
// Empty detect result uses fallback "en" (V2 passes null fallback → "en"), never preferred[0].
// Invalid preferred format returns "" (callers with user input should use DetectLanguageOfStringErr).
func DetectLanguageOfString(text string, preferredVariants []string, detect func(string) string) string {
	code, err := DetectLanguageOfStringErr(text, preferredVariants, detect)
	if err != nil {
		return ""
	}
	return code
}

// DetectLanguageOfStringErr is DetectLanguageOfString with BadRequestError on invalid preferredVariants.
func DetectLanguageOfStringErr(text string, preferredVariants []string, detect func(string) string) (string, error) {
	return DetectLanguageOfStringWithFallback(text, "", preferredVariants, detect, nil)
}

// DetectLanguageOfStringWithFallback ports detectLanguageOfString(text, fallbackLanguage, preferred, …).
// fallbackLanguage empty → "en" when detect yields empty (Java null fallback).
// isKnown optional: when non-nil and preferred matches short code, unknown codes → BadRequestError
// (Java parseLanguage / Languages.getLanguageForShortCode).
func DetectLanguageOfStringWithFallback(
	text, fallbackLanguage string,
	preferredVariants []string,
	detect func(string) string,
	isKnown func(code string) bool,
) (string, error) {
	fn := detect
	if fn == nil {
		fn = commandline.DetectLanguageHeuristic
	}
	code := fn(text)
	if code == "" {
		if fallbackLanguage != "" {
			code = fallbackLanguage
		} else {
			code = "en"
		}
	}

	if len(preferredVariants) > 0 {
		short := languageShortCode(code)
		for _, preferredVariant := range preferredVariants {
			if preferredVariant == "" {
				continue
			}
			// Java: preferredVariant.contains("-")
			if !strings.Contains(preferredVariant, "-") {
				return "", NewBadRequestError(
					"Invalid format for 'preferredVariants', expected a dash as in 'en-GB': '" + preferredVariant + "'")
			}
			// Java: preferredVariant.split("-")[0] — case-sensitive equals lang.getShortCode()
			preferredVariantLang := preferredVariant
			if i := strings.IndexByte(preferredVariant, '-'); i >= 0 {
				preferredVariantLang = preferredVariant[:i]
			}
			if preferredVariantLang == short {
				// Java: lang = parseLanguage(preferredVariant)
				if isKnown != nil && !isKnown(preferredVariant) {
					return "", NewBadRequestError(
						"Invalid 'preferredVariants', no such language/variant found: '" + preferredVariant + "'")
				}
				code = canonicalizePreferredVariant(preferredVariant)
			}
		}
		return code, nil
	}

	// Java: preferred empty → getDefaultLanguageVariant() when non-null (Language default returns this).
	return applyDefaultLanguageVariant(code), nil
}

// languageShortCode ports Language.getShortCode() for a shortCodeWithCountryAndVariant string.
func languageShortCode(code string) string {
	if i := strings.IndexByte(code, '-'); i >= 0 {
		return code[:i]
	}
	return code
}

// applyDefaultLanguageVariant ports Language.getDefaultLanguageVariant() code selection
// for detected short codes (English→en-US, German→de-DE, …; base Language returns this).
// Variants inherit the family override (BritishEnglish extends English → still en-US).
func applyDefaultLanguageVariant(code string) string {
	if code == "" {
		return code
	}
	short := strings.ToLower(languageShortCode(code))
	switch short {
	case "en":
		return language.EnglishDefaultLanguageVariantCode()
	case "de":
		return language.GermanDefaultLanguageVariantCode()
	case "pt":
		return language.PortugueseDefaultLanguageVariantCode()
	case "ca":
		return language.CatalanDefaultLanguageVariantCode()
	case "sr":
		return language.SerbianDefaultLanguageVariantCode()
	case "fr":
		return language.FrenchDefaultLanguageVariantCode()
	case "es":
		return language.SpanishDefaultLanguageVariantCode()
	case "nl":
		return language.DutchDefaultLanguageVariantCode()
	case "ga":
		return language.IrishDefaultLanguageVariantCode()
	default:
		// Java Language.getDefaultLanguageVariant() returns this → keep detected code.
		return code
	}
}

// canonicalizePreferredVariant ports Languages.getLanguageForShortCode casing for common tags:
// language lower, 2-letter region upper (de-at → de-AT, en-gb → en-GB).
func canonicalizePreferredVariant(code string) string {
	parts := strings.Split(code, "-")
	if len(parts) == 0 {
		return code
	}
	parts[0] = strings.ToLower(parts[0])
	for i := 1; i < len(parts); i++ {
		p := parts[i]
		switch len(p) {
		case 2:
			parts[i] = strings.ToUpper(p)
		case 4:
			// script subtag (e.g. Latn)
			r := []rune(p)
			if len(r) == 4 {
				r[0] = unicode.ToUpper(r[0])
				for j := 1; j < 4; j++ {
					r[j] = unicode.ToLower(r[j])
				}
				parts[i] = string(r)
			}
		}
	}
	return strings.Join(parts, "-")
}

// ParsePreferredVariants ports V2TextChecker.getPreferredVariants:
// COMMA_WHITESPACE_PATTERN.split; requires language=auto (or multilingual true).
// Missing preferredVariants key → empty list (not an error).
// Java: !"auto".equals(language) && (multilingual == null || multilingual.equals("false"))
func ParsePreferredVariants(parameters map[string]string) ([]string, error) {
	if parameters == nil {
		return nil, nil
	}
	raw, ok := parameters["preferredVariants"]
	if !ok {
		return nil, nil
	}
	// Java: parameters.get("preferredVariants") != null → split even if empty string
	preferred := splitCommaWhitespace(raw)
	lang := parameters["language"]
	multi, multiOK := parameters["multilingual"]
	// multi missing → null in Java; multi=="false" exact
	if lang != "auto" && (!multiOK || multi == "false") {
		return nil, NewBadRequestError("You specified 'preferredVariants' but you didn't specify 'language=auto'")
	}
	return preferred, nil
}

// ValidateNoopLanguages ports TextChecker: noopLanguages only with language=auto.
// Java: params.get("noopLanguages") != null && !autoDetectLanguage where
// autoDetect = "auto".equals(language). Key present (even empty) counts as set.
func ValidateNoopLanguages(parameters map[string]string) error {
	if parameters == nil {
		return nil
	}
	if _, ok := parameters["noopLanguages"]; !ok {
		return nil
	}
	if parameters["language"] != "auto" {
		return NewBadRequestError("You can specify 'noopLanguages' only when also using 'language=auto'")
	}
	return nil
}
