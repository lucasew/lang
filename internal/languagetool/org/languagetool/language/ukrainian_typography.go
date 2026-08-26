package language

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// Ukrainian quote characters (Ukrainian.getOpening/Closing*).
const (
	ukOpenDouble  = "«"
	ukCloseDouble = "»"
	ukOpenSingle  = "‘"
	ukCloseSingle = "’"
)

// UkrainianIsAdvancedTypographyEnabled ports Ukrainian.isAdvancedTypographyEnabled.
// Java returns false (DISABLED) — do not invent enabling.
func UkrainianIsAdvancedTypographyEnabled() bool { return false }

// UkrainianAdvancedTypography ports Ukrainian.toAdvancedTypography when disabled:
// only suggestion tags are replaced with double quotes (Language base with Enabled=false).
// When callers force-enable, quotes use « » / ‘ ’.
func UkrainianAdvancedTypography(input string) string {
	cfg := languagetool.TypographyConfig{
		Enabled:            UkrainianIsAdvancedTypographyEnabled(),
		OpeningDoubleQuote: ukOpenDouble,
		ClosingDoubleQuote: ukCloseDouble,
		OpeningSingleQuote: ukOpenSingle,
		ClosingSingleQuote: ukCloseSingle,
	}
	return languagetool.ToAdvancedTypography(input, cfg)
}

// UkrainianTypographyConfig returns quote config for optional enable (tests / variants).
func UkrainianTypographyConfig(enabled bool) languagetool.TypographyConfig {
	return languagetool.TypographyConfig{
		Enabled:            enabled,
		OpeningDoubleQuote: ukOpenDouble,
		ClosingDoubleQuote: ukCloseDouble,
		OpeningSingleQuote: ukOpenSingle,
		ClosingSingleQuote: ukCloseSingle,
	}
}
