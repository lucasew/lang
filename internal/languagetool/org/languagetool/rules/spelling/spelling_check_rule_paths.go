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
	out := make([]string, 0, 4)
	if sc != "" {
		out = append(out, sc+CustomSpellingFile)
	}
	out = append(out, GlobalSpellingFile)
	switch sc {
	case "en":
		// AbstractEnglishSpellerRule: + "/en/multiwords.txt"
		out = append(out, "/en/multiwords.txt")
	case "pt":
		out = append(out, "pt/multiwords.txt")
	case "es":
		out = append(out, "es/multiwords.txt")
	case "ca":
		out = append(out, "ca/multiwords.txt", "ca/spelling-special.txt")
	}
	return out
}

// GetLanguageVariantSpellingFileName ports getLanguageVariantSpellingFileName.
// Base Java returns null; EN/DE variants return spelling_en-US.txt etc.
func (r *SpellingCheckRule) GetLanguageVariantSpellingFileName() string {
	if r != nil && r.GetLanguageVariantSpellingFileNameFn != nil {
		return r.GetLanguageVariantSpellingFileNameFn()
	}
	if r == nil {
		return ""
	}
	return LanguageVariantSpellingClasspath(r.LanguageCode)
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
