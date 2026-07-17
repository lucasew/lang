package server

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/corepack"

// DefaultCoreLanguages returns LanguageInfo for all corepack-supported languages.
func DefaultCoreLanguages() []LanguageInfo {
	out := make([]LanguageInfo, 0, len(corepack.Supported))
	for _, s := range corepack.Supported {
		out = append(out, LanguageInfo{Name: s.Name, Code: s.Code})
	}
	return out
}
