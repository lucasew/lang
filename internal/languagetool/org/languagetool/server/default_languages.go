package server

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// DefaultCoreLanguages ports ApiV2.getLanguages JSON language list:
//
//  1. Languages.get() → {name, code=getShortCode(), longCode=getShortCodeWithCountryAndVariant()}
//  2. Languages.getLongCodeToLangMapping() entries whose longCode is not already listed
//     (LibreOffice 7.4 fr-FR → fr mappings).
//
// No invent soft variant tables — registry is the Java language-module list.
func DefaultCoreLanguages() []LanguageInfo {
	languagetool.EnsureBuiltInLanguagesRegistered()
	langs := languagetool.GlobalLanguages.Get()
	out := make([]LanguageInfo, 0, len(langs)+16)
	longCodes := map[string]struct{}{}
	for _, lang := range langs {
		long := lang.GetShortCodeWithCountryAndVariant()
		out = append(out, LanguageInfo{
			Name:     lang.GetName(),
			Code:     lang.GetShortCode(),
			LongCode: long,
		})
		longCodes[long] = struct{}{}
	}
	// LibreOffice long-code mappings (Java getLongCodeToLangMapping)
	for longCode, lang := range languagetool.GlobalLanguages.GetLongCodeToLangMapping() {
		if _, ok := longCodes[longCode]; ok {
			continue
		}
		out = append(out, LanguageInfo{
			Name:     lang.GetName(),
			Code:     lang.GetShortCode(),
			LongCode: longCode,
		})
		longCodes[longCode] = struct{}{}
	}
	return out
}
