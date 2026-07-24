package language

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Catalan quote characters (Catalan.getOpening/Closing*).
const (
	caOpenDouble  = "«"
	caCloseDouble = "»"
	caOpenSingle  = "‘"
	caCloseSingle = "’"
)

// Catalan apostrophe + quote special cases after base Language typography.
var (
	// PATTERN_1: (\b[lmnstdLMNSTD])' → $1’
	caTypo1 = regexp.MustCompile(`(\b[lmnstdLMNSTD])'`)
	// PATTERN_2: (\b[lmnstdLMNSTD])’" → $1’«
	caTypo2 = regexp.MustCompile(`(\b[lmnstdLMNSTD])’"`)
	// PATTERN_3: (\b[lmnstdLMNSTD])’' → $1’‘
	caTypo3 = regexp.MustCompile(`(\b[lmnstdLMNSTD])’'`)
)

// CatalanAdvancedTypography ports Catalan.toAdvancedTypography.
func CatalanAdvancedTypography(input string) string {
	cfg := languagetool.TypographyConfig{
		Enabled:            true,
		OpeningDoubleQuote: caOpenDouble,
		ClosingDoubleQuote: caCloseDouble,
		OpeningSingleQuote: caOpenSingle,
		ClosingSingleQuote: caCloseSingle,
	}
	out := languagetool.ToAdvancedTypography(input, cfg)
	// special cases: apostrophe + quotation marks
	out = caTypo1.ReplaceAllString(out, "$1’")
	out = caTypo2.ReplaceAllString(out, "$1’"+caOpenDouble)
	out = caTypo3.ReplaceAllString(out, "$1’"+caOpenSingle)
	return out
}

// CatalanIsAdvancedTypographyEnabled ports Catalan.isAdvancedTypographyEnabled (true).
func CatalanIsAdvancedTypographyEnabled() bool { return true }

// CatalanHasMinMatchesRules ports Catalan.hasMinMatchesRules (true).
func CatalanHasMinMatchesRules() bool { return true }

// GetOpeningDoubleQuote ports Catalan.getOpeningDoubleQuote ("«").
func (v CatalanVariant) GetOpeningDoubleQuote() string { return caOpenDouble }

// GetClosingDoubleQuote ports Catalan.getClosingDoubleQuote ("»").
func (v CatalanVariant) GetClosingDoubleQuote() string { return caCloseDouble }

// GetOpeningSingleQuote ports Catalan.getOpeningSingleQuote ("‘").
func (v CatalanVariant) GetOpeningSingleQuote() string { return caOpenSingle }

// GetClosingSingleQuote ports Catalan.getClosingSingleQuote ("’").
func (v CatalanVariant) GetClosingSingleQuote() string { return caCloseSingle }

// IsAdvancedTypographyEnabled ports Catalan.isAdvancedTypographyEnabled (true).
func (v CatalanVariant) IsAdvancedTypographyEnabled() bool { return true }

// HasMinMatchesRules ports Catalan.hasMinMatchesRules (true).
func (v CatalanVariant) HasMinMatchesRules() bool { return true }

// ToAdvancedTypography ports Catalan.toAdvancedTypography.
func (v CatalanVariant) ToAdvancedTypography(input string) string {
	return CatalanAdvancedTypography(input)
}

// GetDefaultRulePriorityForStyle ports Catalan.getDefaultRulePriorityForStyle (-50).
func (v CatalanVariant) GetDefaultRulePriorityForStyle() int { return -50 }
