package detector

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// CanLanguageBeDetected ports LanguageIdentifierService.canLanguageBeDetected
// for FastTextDetector / NGramDetector (detector must not import identifier).
// Java: Languages.isLanguageSupported(langCode) || additionalLanguageCodes.contains(langCode)
func CanLanguageBeDetected(langCode string, additional []string) bool {
	if langCode == "" {
		return false
	}
	// Java Languages class is initialized before any detect call.
	languagetool.EnsureBuiltInLanguagesRegistered()
	for _, a := range additional {
		if a == langCode {
			return true
		}
	}
	// Invalid format → false (Java isLanguageSupported throws; we fail closed).
	if err := languagetool.ValidateLanguageCodeFormat(langCode); err != nil {
		return false
	}
	return languagetool.GlobalLanguages.IsLanguageSupported(langCode)
}
