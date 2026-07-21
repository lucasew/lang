package server

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/commandline"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
)

// DetectLanguageResult ports TextChecker.detectLanguageOfString return surface:
// resolved language code + detection confidence/source from the identifier.
// Java: new DetectedLanguage(null, lang, detected != null ? conf : 0f, source).
type DetectLanguageResult struct {
	Code       string
	Confidence float32
	Source     *string
}

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
	r, err := DetectLanguageOfStringWithFallback(text, "", preferredVariants, detect, nil)
	return r.Code, err
}

// DetectLanguageOfStringResult is DetectLanguageOfStringErr with confidence/source.
func DetectLanguageOfStringResult(text string, preferredVariants []string, detect func(string) string) (DetectLanguageResult, error) {
	return DetectLanguageOfStringWithFallback(text, "", preferredVariants, detect, nil)
}

// DetectLanguageOfStringWithFallback ports detectLanguageOfString(text, fallbackLanguage, preferred, …).
// fallbackLanguage empty → "en" when detect yields empty (Java null fallback).
// isKnown optional: when non-nil and preferred matches short code, unknown codes → BadRequestError
// (Java parseLanguage / Languages.getLanguageForShortCode).
//
// Confidence: Java uses detected.getDetectionConfidence() when identifier returns non-null,
// else 0f. String inject without conf → 0 when empty detect, else 0 (no invent 0.5).
// Prefer DetectLanguageOfStringFromDetected when identifier.DetectedLanguage is available.
func DetectLanguageOfStringWithFallback(
	text, fallbackLanguage string,
	preferredVariants []string,
	detect func(string) string,
	isKnown func(code string) bool,
) (DetectLanguageResult, error) {
	fn := detect
	if fn == nil {
		fn = commandline.DetectLanguageHeuristic
	}
	raw := fn(text)
	// Java: detected == null → conf 0f; else keep identifier confidence (unknown for string inject → 0)
	var conf float32
	var src *string
	hadDetect := raw != ""
	code := raw
	if code == "" {
		if fallbackLanguage != "" {
			code = fallbackLanguage
		} else {
			code = "en"
		}
		conf = 0
	} else {
		// String-only inject has no identifier confidence — do not invent 0.5.
		conf = 0
	}
	_ = hadDetect

	if len(preferredVariants) > 0 {
		short := languageShortCode(code)
		for _, preferredVariant := range preferredVariants {
			if preferredVariant == "" {
				continue
			}
			// Java: preferredVariant.contains("-")
			if !strings.Contains(preferredVariant, "-") {
				return DetectLanguageResult{}, NewBadRequestError(
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
					return DetectLanguageResult{}, NewBadRequestError(
						"Invalid 'preferredVariants', no such language/variant found: '" + preferredVariant + "'")
				}
				code = canonicalizePreferredVariant(preferredVariant)
			}
		}
		return DetectLanguageResult{Code: code, Confidence: conf, Source: src}, nil
	}

	// Java: preferred empty → getDefaultLanguageVariant() when non-null (Language default returns this).
	return DetectLanguageResult{Code: applyDefaultLanguageVariant(code), Confidence: conf, Source: src}, nil
}

// DetectLanguageOfStringFromDetected applies preferred/default variant resolution to an
// identifier DetectedLanguage (preserves confidence/source). nil detected → fallback "en", conf 0.
func DetectLanguageOfStringFromDetected(
	detected *languagetool.DetectedLanguage,
	fallbackLanguage string,
	preferredVariants []string,
	isKnown func(code string) bool,
) (DetectLanguageResult, error) {
	var code string
	var conf float32
	var src *string
	if detected == nil {
		if fallbackLanguage != "" {
			code = fallbackLanguage
		} else {
			code = "en"
		}
		conf = 0
	} else {
		code = detected.GetDetectedLanguageCode()
		conf = detected.GetDetectionConfidence()
		src = detected.GetDetectionSource()
	}

	if len(preferredVariants) > 0 {
		short := languageShortCode(code)
		for _, preferredVariant := range preferredVariants {
			if preferredVariant == "" {
				continue
			}
			if !strings.Contains(preferredVariant, "-") {
				return DetectLanguageResult{}, NewBadRequestError(
					"Invalid format for 'preferredVariants', expected a dash as in 'en-GB': '" + preferredVariant + "'")
			}
			preferredVariantLang := preferredVariant
			if i := strings.IndexByte(preferredVariant, '-'); i >= 0 {
				preferredVariantLang = preferredVariant[:i]
			}
			if preferredVariantLang == short {
				if isKnown != nil && !isKnown(preferredVariant) {
					return DetectLanguageResult{}, NewBadRequestError(
						"Invalid 'preferredVariants', no such language/variant found: '" + preferredVariant + "'")
				}
				code = canonicalizePreferredVariant(preferredVariant)
			}
		}
		return DetectLanguageResult{Code: code, Confidence: conf, Source: src}, nil
	}
	return DetectLanguageResult{Code: applyDefaultLanguageVariant(code), Confidence: conf, Source: src}, nil
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

// ParseNoopLanguages ports TextChecker:
// params.get("noopLanguages") != null ? Arrays.asList(split(",")) : emptyList.
func ParseNoopLanguages(parameters map[string]string) []string {
	return parseOptionalCommaParam(parameters, "noopLanguages")
}

// ParsePreferredLanguages ports TextChecker preferredLanguages split(",").
func ParsePreferredLanguages(parameters map[string]string) []string {
	return parseOptionalCommaParam(parameters, "preferredLanguages")
}

func parseOptionalCommaParam(parameters map[string]string, key string) []string {
	if parameters == nil {
		return nil
	}
	if _, ok := parameters[key]; !ok {
		return nil
	}
	return commaSeparated(parameters[key])
}

// DetectLanguageOfString ports TextChecker.detectLanguageOfString using the
// instance languageIdentifier (cleanAndShorten + detect + preferred/default variants).
// nil receiver / nil identifier falls back to package DetectLanguageOfStringResult (heuristic).
func (t *TextChecker) DetectLanguageOfString(
	text string,
	preferredVariants, noopLangs, preferredLangs []string,
) (DetectLanguageResult, error) {
	if t == nil || t.LanguageIdentifier == nil {
		return DetectLanguageOfStringResult(text, preferredVariants, nil)
	}
	// Java: cleanText = languageIdentifier.cleanAndShortenText(text)
	//       detected = languageIdentifier.detectLanguage(cleanText, noopLangs, preferredLangs, forcePreferred)
	clean := t.LanguageIdentifier.CleanAndShortenText(text)
	if noopLangs == nil {
		noopLangs = []string{}
	}
	if preferredLangs == nil {
		preferredLangs = []string{}
	}
	detected := t.LanguageIdentifier.Detect(clean, noopLangs, preferredLangs)
	return DetectLanguageOfStringFromDetected(detected, "", preferredVariants, nil)
}
