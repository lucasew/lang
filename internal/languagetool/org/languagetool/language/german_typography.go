package language

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// German quote characters (German.getOpening/Closing*).
const (
	deOpenDouble  = "„"
	deCloseDouble = "“"
	deOpenSingle  = "‚"
	deCloseSingle = "‘"
)

// germanTypographyPattern ports German.TYPOGRAPHY_PATTERN:
// \b([a-zA-Z]\.)([a-zA-Z]\.) → non-breaking space between abbreviation dots (z.B. → z.\u00a0B.).
var germanTypographyPattern = regexp.MustCompile(`\b([a-zA-Z]\.)([a-zA-Z]\.)`)

// ToAdvancedTypography ports German.toAdvancedTypography (DE/AT) or
// SwissGerman quotes (CH: « »). TYPOGRAPHY_PATTERN applied twice after base.
func (v GermanVariant) ToAdvancedTypography(input string) string {
	if strings.EqualFold(v.ShortCode, "de-CH") || strings.HasSuffix(strings.ToUpper(v.ShortCode), "-CH") {
		return SwissGermanAdvancedTypography(input)
	}
	return GermanAdvancedTypography(input)
}

// GermanAdvancedTypography is the package-level entry used by twins.
func GermanAdvancedTypography(input string) string {
	cfg := languagetool.TypographyConfig{
		Enabled:            true,
		OpeningDoubleQuote: deOpenDouble,
		ClosingDoubleQuote: deCloseDouble,
		OpeningSingleQuote: deOpenSingle,
		ClosingSingleQuote: deCloseSingle,
	}
	out := languagetool.ToAdvancedTypography(input, cfg)
	// Java applies the same replace twice (i.d.R. needs two passes).
	out = germanTypographyPattern.ReplaceAllString(out, "$1\u00a0$2")
	out = germanTypographyPattern.ReplaceAllString(out, "$1\u00a0$2")
	return out
}

// IsAdvancedTypographyEnabled ports German.isAdvancedTypographyEnabled (true).
func (v GermanVariant) IsAdvancedTypographyEnabled() bool { return true }
