package language

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// Simple advanced-typography languages: base Language.toAdvancedTypography with
// language-specific quote characters and isAdvancedTypographyEnabled=true.
// No extra post-processing (unlike DE TYPOGRAPHY_PATTERN, FR spaces, CA apostrophe+quote).

func simpleAdvancedTypography(input, openD, closeD, openS, closeS string) string {
	cfg := languagetool.TypographyConfig{
		Enabled:            true,
		OpeningDoubleQuote: openD,
		ClosingDoubleQuote: closeD,
		OpeningSingleQuote: openS,
		ClosingSingleQuote: closeS,
	}
	return languagetool.ToAdvancedTypography(input, cfg)
}

// Spanish: « » ‘ ’
func SpanishAdvancedTypography(input string) string {
	return simpleAdvancedTypography(input, "«", "»", "‘", "’")
}
func SpanishIsAdvancedTypographyEnabled() bool { return true }

// Dutch: “ ” ‘ ’
func DutchAdvancedTypography(input string) string {
	return simpleAdvancedTypography(input, "“", "”", "‘", "’")
}
func DutchIsAdvancedTypographyEnabled() bool { return true }

// Portuguese (base / non-Portugal locales): “ ” ‘ ’
// Java Portuguese.getOpening/ClosingDoubleQuote.
func PortugueseAdvancedTypography(input string) string {
	return simpleAdvancedTypography(input, "“", "”", "‘", "’")
}
func PortugueseIsAdvancedTypographyEnabled() bool { return true }

// PortugalPortugueseAdvancedTypography ports PortugalPortuguese double quotes « ».
// Singles stay base Portuguese ‘ ’.
func PortugalPortugueseAdvancedTypography(input string) string {
	return simpleAdvancedTypography(input, "«", "»", "‘", "’")
}

// Russian: « » ‘ ’
func RussianAdvancedTypography(input string) string {
	return simpleAdvancedTypography(input, "«", "»", "‘", "’")
}
func RussianIsAdvancedTypographyEnabled() bool { return true }

// Belarusian: « » ‘ ’ (Belarusian.getOpening/Closing* + isAdvancedTypographyEnabled=true).
func BelarusianAdvancedTypography(input string) string {
	return simpleAdvancedTypography(input, "«", "»", "‘", "’")
}
func BelarusianIsAdvancedTypographyEnabled() bool { return true }

// English: “ ” ‘ ’
func EnglishAdvancedTypography(input string) string {
	return simpleAdvancedTypography(input, "“", "”", "‘", "’")
}
func EnglishIsAdvancedTypographyEnabled() bool { return true }
