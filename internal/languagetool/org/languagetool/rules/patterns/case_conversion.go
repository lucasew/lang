package patterns

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CaseConversion ports Match.CaseConversion.
type CaseConversion string

const (
	CaseNone       CaseConversion = "NONE"
	CaseStartLower CaseConversion = "STARTLOWER"
	CaseStartUpper CaseConversion = "STARTUPPER"
	CaseAllLower   CaseConversion = "ALLLOWER"
	CaseAllUpper   CaseConversion = "ALLUPPER"
	CasePreserve   CaseConversion = "PRESERVE"
	CaseFirstUpper CaseConversion = "FIRSTUPPER"
	CaseNoTashkeel CaseConversion = "NOTASHKEEL"
)

// IncludeRange ports Match.IncludeRange.
type IncludeRange string

const (
	IncludeNone      IncludeRange = "NONE"
	IncludeFollowing IncludeRange = "FOLLOWING"
	IncludeAll       IncludeRange = "ALL"
)

// allCaseConversions is the full Java Match.CaseConversion enum.
var allCaseConversions = []CaseConversion{
	CaseNone, CaseStartLower, CaseStartUpper, CaseAllLower, CaseAllUpper,
	CasePreserve, CaseFirstUpper, CaseNoTashkeel,
}

// ParseCaseConversion ports Match.CaseConversion.valueOf (upper-case names).
// Unknown names return false (no invent); Java would throw.
func ParseCaseConversion(name string) (CaseConversion, bool) {
	u := strings.ToUpper(tools.JavaStringTrim(name))
	for _, c := range allCaseConversions {
		if string(c) == u {
			return c, true
		}
	}
	return CaseNone, false
}

// allIncludeRanges is the full Java Match.IncludeRange enum.
var allIncludeRanges = []IncludeRange{IncludeNone, IncludeFollowing, IncludeAll}

// ParseIncludeRange ports Match.IncludeRange.valueOf (upper-case names).
func ParseIncludeRange(name string) (IncludeRange, bool) {
	u := strings.ToUpper(tools.JavaStringTrim(name))
	for _, r := range allIncludeRanges {
		if string(r) == u {
			return r, true
		}
	}
	return IncludeNone, false
}

// ConvertCase ports CaseConversionHelper.convertCase without language-specific rules.
func ConvertCase(conversion CaseConversion, s, sample string) string {
	return ConvertCaseLang(conversion, s, sample, "")
}

// ConvertCaseLang ports CaseConversionHelper.convertCase with language short code
// (Dutch "ij" → "IJ" when uppercasing the first char).
func ConvertCaseLang(conversion CaseConversion, s, sample, langShortCode string) string {
	if tools.IsEmptyStr(s) {
		return s
	}
	token := s
	switch conversion {
	case CaseNone:
		// no-op
	case CasePreserve:
		if tools.StartsWithUppercase(sample) {
			if tools.IsAllUppercase(sample) {
				token = strings.ToUpper(token)
			} else {
				token = uppercaseFirstCharLang(token, langShortCode)
			}
		}
	case CaseStartLower:
		rs := []rune(token)
		if len(rs) > 0 {
			rs[0] = unicode.ToLower(rs[0])
			token = string(rs)
		}
	case CaseStartUpper:
		token = uppercaseFirstCharLang(token, langShortCode)
	case CaseAllUpper:
		token = strings.ToUpper(token)
	case CaseFirstUpper:
		token = uppercaseFirstCharLang(strings.ToLower(token), langShortCode)
	case CaseAllLower:
		token = strings.ToLower(token)
	case CaseNoTashkeel:
		token = tools.RemoveTashkeel(token)
	}
	return token
}

func uppercaseFirstCharLang(str, langShortCode string) string {
	return tools.UppercaseFirstCharLang(str, langShortCode)
}
