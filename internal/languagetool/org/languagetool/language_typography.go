package languagetool

import (
	"regexp"
	"strings"
)

// Quote marks used by Language typography (defaults match Java Language base).
const (
	DefaultOpeningDoubleQuote = "“"
	DefaultClosingDoubleQuote = "”"
	DefaultOpeningSingleQuote = "‘"
	DefaultClosingSingleQuote = "’"
	suggestionOpenTag         = "<suggestion>"
	suggestionCloseTag        = "</suggestion>"
)

var (
	insideSuggestionRE = regexp.MustCompile(`(?s)<suggestion>(.+?)</suggestion>`)
	nbSpace1RE         = regexp.MustCompile(`\b([a-zA-Z]\.) ([a-zA-Z]\.)`)
	nbSpace2RE         = regexp.MustCompile(`\b([a-zA-Z]\.) `)
	apostropheRE       = regexp.MustCompile(`([\p{L}\d-])'([\p{L}«])`)
	quotedCharRE       = regexp.MustCompile(` '(.)'`)
	typog1RE           = regexp.MustCompile(`([\x{202f}\x{00a0} «"\(])'`)
	typog2RE           = regexp.MustCompile(`'([\x{202f}\x{00a0} !\?,\.;:"\)])`)
	typog3RE           = regexp.MustCompile(`‘s\b([^’])`)
	typog4RE           = regexp.MustCompile(`([ \(])"`)
	typog5RE           = regexp.MustCompile(`"([\x{202f}\x{00a0} !\?,\.;:\)])`)
)

// TypographyConfig controls Language.toAdvancedTypography quote characters.
type TypographyConfig struct {
	Enabled            bool
	OpeningDoubleQuote string
	ClosingDoubleQuote string
	OpeningSingleQuote string
	ClosingSingleQuote string
}

func DefaultTypographyConfig() TypographyConfig {
	return TypographyConfig{
		OpeningDoubleQuote: DefaultOpeningDoubleQuote,
		ClosingDoubleQuote: DefaultClosingDoubleQuote,
		OpeningSingleQuote: DefaultOpeningSingleQuote,
		ClosingSingleQuote: DefaultClosingSingleQuote,
	}
}

// ToAdvancedTypography ports Language.toAdvancedTypography.
// When Enabled is false, only suggestion tags are replaced with double quotes.
func ToAdvancedTypography(input string, cfg TypographyConfig) string {
	openD := cfg.OpeningDoubleQuote
	closeD := cfg.ClosingDoubleQuote
	if openD == "" {
		openD = DefaultOpeningDoubleQuote
	}
	if closeD == "" {
		closeD = DefaultClosingDoubleQuote
	}
	openS := cfg.OpeningSingleQuote
	closeS := cfg.ClosingSingleQuote
	if openS == "" {
		openS = DefaultOpeningSingleQuote
	}
	if closeS == "" {
		closeS = DefaultClosingSingleQuote
	}

	if !cfg.Enabled {
		return strings.ReplaceAll(strings.ReplaceAll(input, suggestionOpenTag, openD), suggestionCloseTag, closeD)
	}

	output := input
	// Preserve content inside <suggestion>...</suggestion>
	var preserved []string
	for {
		loc := insideSuggestionRE.FindStringSubmatchIndex(output)
		if loc == nil {
			break
		}
		group := output[loc[2]:loc[3]]
		preserved = append(preserved, group)
		placeholder := `\` + itoa(len(preserved)-1)
		// replace first occurrence of full match
		full := output[loc[0]:loc[1]]
		output = strings.Replace(output, full, placeholder, 1)
	}

	output = strings.ReplaceAll(output, "...", "…")
	output = nbSpace1RE.ReplaceAllString(output, "$1\u00a0$2")
	output = nbSpace2RE.ReplaceAllString(output, "$1\u00a0")
	output = apostropheRE.ReplaceAllString(output, "$1’$2")

	if strings.HasPrefix(output, "'") {
		output = openS + output[1:]
	}
	if strings.HasSuffix(output, "'") {
		output = output[:len(output)-1] + closeS
	}
	output = quotedCharRE.ReplaceAllString(output, " "+openS+"$1"+closeS)
	output = typog1RE.ReplaceAllString(output, "$1"+openS)
	output = typog2RE.ReplaceAllString(output, closeS+"$1")
	output = typog3RE.ReplaceAllString(output, "’s$1")

	if strings.HasPrefix(output, `"`) {
		output = openD + output[1:]
	}
	if strings.HasSuffix(output, `"`) {
		output = output[:len(output)-1] + closeD
	}
	output = typog4RE.ReplaceAllString(output, "$1"+openD)
	output = typog5RE.ReplaceAllString(output, closeD+"$1")

	for i, s := range preserved {
		output = strings.Replace(output, `\`+itoa(i), openD+s+closeD, 1)
	}
	output = strings.ReplaceAll(output, suggestionOpenTag, openD)
	output = strings.ReplaceAll(output, suggestionCloseTag, closeD)
	return output
}

// ToCommonTypography enables advanced typography with default quotes.
func ToCommonTypography(input string) string {
	cfg := DefaultTypographyConfig()
	cfg.Enabled = true
	return ToAdvancedTypography(input, cfg)
}

// AdaptSuggestion ports Language.adaptSuggestion (identity by default).
func AdaptSuggestion(s, originalErrorStr string) string {
	_ = originalErrorStr
	return s
}

// AdaptSuggestionsList ports Language.adaptSuggestionsList.
func AdaptSuggestionsList(suggestions []string, originalErrorStr string) []string {
	out := make([]string, 0, len(suggestions))
	for _, s := range suggestions {
		out = append(out, AdaptSuggestion(s, originalErrorStr))
	}
	return out
}

// EqualsConsiderVariantsIfSpecified ports Language.equalsConsiderVariantsIfSpecified.
// If either code lacks a variant (no '-'), compare short codes only.
func EqualsConsiderVariantsIfSpecified(a, b string) bool {
	if a == b {
		return true
	}
	as, bs := shortLang(a), shortLang(b)
	aHasVar := strings.Contains(a, "-") || strings.Contains(a, "_")
	bHasVar := strings.Contains(b, "-") || strings.Contains(b, "_")
	if !aHasVar || !bHasVar {
		return as == bs
	}
	return normalizeLangCode(a) == normalizeLangCode(b)
}

func shortLang(code string) string {
	code = strings.ReplaceAll(code, "_", "-")
	if i := strings.IndexByte(code, '-'); i >= 0 {
		return code[:i]
	}
	return code
}

func normalizeLangCode(code string) string {
	return strings.ReplaceAll(code, "_", "-")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
