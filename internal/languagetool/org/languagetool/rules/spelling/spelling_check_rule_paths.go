package spelling

import "strings"

// Path constants matching SpellingCheckRule private static finals
// (prefix is language short code; GLOBAL_SPELLING_FILE has no lang prefix).
const (
	// SpellingProhibitFile ports "/hunspell/prohibit.txt".
	SpellingProhibitFile = "/hunspell/prohibit.txt"
	// CustomSpellingProhibitFile ports "/hunspell/prohibit_custom.txt".
	CustomSpellingProhibitFile = "/hunspell/prohibit_custom.txt"
)

// ShortCode ports language.getShortCode() for resource path construction.
func (r *SpellingCheckRule) ShortCode() string {
	if r == nil {
		return ""
	}
	lang := r.LanguageCode
	if i := strings.IndexByte(lang, '-'); i > 0 {
		lang = lang[:i]
	}
	if i := strings.IndexByte(lang, '_'); i > 0 {
		lang = lang[:i]
	}
	return lang
}

// GetIgnoreFileName ports getIgnoreFileName:
// language.getShortCode() + SPELLING_IGNORE_FILE ("/hunspell/ignore.txt").
func (r *SpellingCheckRule) GetIgnoreFileName() string {
	if r != nil && r.GetIgnoreFileNameFn != nil {
		return r.GetIgnoreFileNameFn()
	}
	if r == nil || r.ShortCode() == "" {
		return ""
	}
	return r.ShortCode() + SpellingIgnoreFile
}

// GetSpellingFileName ports getSpellingFileName:
// language.getShortCode() + SPELLING_FILE ("/hunspell/spelling.txt").
// May return empty when overridden to null (rare).
func (r *SpellingCheckRule) GetSpellingFileName() string {
	if r != nil && r.GetSpellingFileNameFn != nil {
		return r.GetSpellingFileNameFn()
	}
	if r == nil || r.ShortCode() == "" {
		return ""
	}
	return r.ShortCode() + SpellingFile
}

// GetAdditionalSpellingFileNames ports getAdditionalSpellingFileNames:
// [short+CUSTOM_SPELLING_FILE, GLOBAL_SPELLING_FILE] plus language extras
// (EN multiwords; PT/ES multiwords; CA multiwords + spelling-special).
func (r *SpellingCheckRule) GetAdditionalSpellingFileNames() []string {
	if r != nil && r.GetAdditionalSpellingFileNamesFn != nil {
		return r.GetAdditionalSpellingFileNamesFn()
	}
	if r == nil {
		return nil
	}
	sc := r.ShortCode()
	// Default Java SpellingCheckRule.getAdditionalSpellingFileNames:
	// [short+CUSTOM_SPELLING_FILE, GLOBAL_SPELLING_FILE]
	// Language overrides replace/extend this list entirely when they override the method.
	switch sc {
	case "en":
		// AbstractEnglishSpellerRule
		return []string{sc + CustomSpellingFile, GlobalSpellingFile, "/en/multiwords.txt"}
	case "pt":
		// MorfologikPortugueseSpellerRule: GLOBAL, pt/spelling.txt, multiwords (no hunspell custom)
		return []string{GlobalSpellingFile, "pt/spelling.txt", "pt/multiwords.txt"}
	case "es":
		// MorfologikSpanishSpellerRule: "/es/"+CUSTOM, GLOBAL, multiwords
		// Java: "/es/" + "/hunspell/spelling_custom.txt" → "/es//hunspell/..."
		return []string{"/es/" + CustomSpellingFile, GlobalSpellingFile, "/es/multiwords.txt"}
	case "ca":
		// MorfologikCatalanSpellerRule: "/ca/"+CUSTOM, GLOBAL, multiwords, spelling-special
		// Java string concat: "/ca/" + "/hunspell/spelling_custom.txt" → "/ca//hunspell/..."
		return []string{"/ca/" + CustomSpellingFile, GlobalSpellingFile, "/ca/multiwords.txt", "/ca/spelling-special.txt"}
	default:
		if sc == "" {
			return []string{GlobalSpellingFile}
		}
		return []string{sc + CustomSpellingFile, GlobalSpellingFile}
	}
}

// GetLanguageVariantSpellingFileName ports getLanguageVariantSpellingFileName.
// Base Java returns null (SPELLING_FILE_VARIANT); EN/DE variants override via Fn.
// When Fn is unset, LanguageVariantSpellingClasspath maps known EN/DE-AT/CH codes
// (same paths Java subclasses return) so bare SpellingCheckRule{LanguageCode: "en-US"}
// still discovers the variant accept list; empty LanguageCode → empty (Java null).
func (r *SpellingCheckRule) GetLanguageVariantSpellingFileName() string {
	if r != nil && r.GetLanguageVariantSpellingFileNameFn != nil {
		return r.GetLanguageVariantSpellingFileNameFn()
	}
	if r == nil {
		return ""
	}
	return LanguageVariantSpellingClasspath(r.LanguageCode)
}

// PlainTextSpellingFileNames ports MorfologikSpellerRule.initSpeller plainTextDicts
// path composition before resourceExists:
//
//	if getSpellingFileName() != null → add
//	for getAdditionalSpellingFileNames() → add
//
// Language-variant file is separate (languageVariantPlainTextDict).
func (r *SpellingCheckRule) PlainTextSpellingFileNames() []string {
	if r == nil {
		return nil
	}
	var out []string
	if name := r.GetSpellingFileName(); name != "" {
		out = append(out, name)
	}
	out = append(out, r.GetAdditionalSpellingFileNames()...)
	return out
}

// CollectExistingPlainTextSpellingFileNames ports initSpeller resourceExists filter
// over PlainTextSpellingFileNames. Missing resources are omitted (fail-closed).
func (r *SpellingCheckRule) CollectExistingPlainTextSpellingFileNames() []string {
	names := r.PlainTextSpellingFileNames()
	if len(names) == 0 {
		return nil
	}
	var out []string
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if DiscoverSpellingResource(strings.TrimPrefix(name, "/")) == "" {
			continue
		}
		out = append(out, name)
	}
	return out
}

// GetProhibitFileName ports getProhibitFileName:
// language.getShortCode() + "/hunspell/prohibit.txt".
func (r *SpellingCheckRule) GetProhibitFileName() string {
	if r != nil && r.GetProhibitFileNameFn != nil {
		return r.GetProhibitFileNameFn()
	}
	if r == nil || r.ShortCode() == "" {
		return ""
	}
	return r.ShortCode() + SpellingProhibitFile
}

// GetAdditionalProhibitFileNames ports getAdditionalProhibitFileNames:
// singleton short + "/hunspell/prohibit_custom.txt".
func (r *SpellingCheckRule) GetAdditionalProhibitFileNames() []string {
	if r != nil && r.GetAdditionalProhibitFileNamesFn != nil {
		return r.GetAdditionalProhibitFileNamesFn()
	}
	if r == nil || r.ShortCode() == "" {
		return nil
	}
	return []string{r.ShortCode() + CustomSpellingProhibitFile}
}
