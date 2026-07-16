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

// ConvertCase ports CaseConversionHelper.convertCase.
// lang is unused until language-specific uppercase rules are wired; uppercase uses tools helpers.
func ConvertCase(conversion CaseConversion, s, sample string) string {
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
				token = tools.UppercaseFirstChar(token)
			}
		}
	case CaseStartLower:
		rs := []rune(token)
		if len(rs) > 0 {
			rs[0] = unicode.ToLower(rs[0])
			token = string(rs)
		}
	case CaseStartUpper:
		token = tools.UppercaseFirstChar(token)
	case CaseAllUpper:
		token = strings.ToUpper(token)
	case CaseFirstUpper:
		token = tools.UppercaseFirstChar(strings.ToLower(token))
	case CaseAllLower:
		token = strings.ToLower(token)
	case CaseNoTashkeel:
		token = tools.RemoveTashkeel(token)
	}
	return token
}
