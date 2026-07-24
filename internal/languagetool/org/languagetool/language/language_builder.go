package language

import (
	"path/filepath"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// LanguageMetaFromFile extends LanguageMeta with the rules XML path.
type LanguageMetaFromFile struct {
	languagetool.LanguageMeta
	RulesFile  string
	Additional bool // Java isExternal for additional languages
}

// ExtendedLanguage ports LanguageBuilder.ExtendedLanguage — base language + extra rule file.
type ExtendedLanguage struct {
	BaseCode   string // short code of base language
	Name       string // display name from filename
	RuleFile   string
	BaseRules  []string // base language rule file names when known
}

// GetName ports ExtendedLanguage.getName.
func (e ExtendedLanguage) GetName() string { return e.Name }

// GetShortCode ports getShortCode from base.
func (e ExtendedLanguage) GetShortCode() string {
	code := e.BaseCode
	if i := strings.IndexByte(code, '_'); i >= 0 {
		return code[:i]
	}
	if i := strings.IndexByte(code, '-'); i >= 0 {
		return code[:i]
	}
	return code
}

// GetRuleFileNames ports getRuleFileNames — base rules + absolute rule file.
func (e ExtendedLanguage) GetRuleFileNames() []string {
	out := append([]string(nil), e.BaseRules...)
	out = append(out, e.RuleFile)
	return out
}

// IsExternal ports isExternal → true for ExtendedLanguage.
func (e ExtendedLanguage) IsExternal() bool { return true }

// MakeAdditionalLanguage ports LanguageBuilder.makeAdditionalLanguage.
// File name must be rules-<code>-<Name>.xml with exactly 3 dash-separated parts
// (Java: parts.length == 3).
func MakeAdditionalLanguage(filename string) (LanguageMetaFromFile, error) {
	base := filepath.Base(filename)
	if !strings.HasSuffix(base, ".xml") {
		return LanguageMetaFromFile{}, NewRuleFilenameException(filename)
	}
	parts := strings.Split(base, "-")
	// Java: startsWithRules && parts.length == 3 && code length 2, 3 (ast), or 5 (en_US)
	if len(parts) != 3 || parts[0] != "rules" {
		return LanguageMetaFromFile{}, NewRuleFilenameException(filename)
	}
	code := parts[1]
	okLen := len(code) == 2 || len(code) == 3 || len(code) == 5
	if !okLen {
		return LanguageMetaFromFile{}, NewRuleFilenameException(filename)
	}
	namePart := strings.TrimSuffix(parts[2], ".xml")
	meta := LanguageMetaFromFile{
		LanguageMeta: languagetool.LanguageMeta{
			Name: namePart,
			Code: code,
		},
		RulesFile:  filename,
		Additional: true,
	}
	if !languagetool.GlobalLanguages.IsLanguageSupported(code) {
		languagetool.GlobalLanguages.Register(meta.LanguageMeta)
	}
	return meta, nil
}

// MakeExtendedLanguage builds ExtendedLanguage when base code is already supported.
func MakeExtendedLanguage(filename string, baseRuleFiles []string) (ExtendedLanguage, error) {
	meta, err := MakeAdditionalLanguage(filename)
	if err != nil {
		return ExtendedLanguage{}, err
	}
	return ExtendedLanguage{
		BaseCode:  meta.Code,
		Name:      meta.Name,
		RuleFile:  filename,
		BaseRules: append([]string(nil), baseRuleFiles...),
	}, nil
}

// ShortCodeFromParts ports anonymous Language.getShortCode for unsupported codes.
func ShortCodeFromParts(codePart string) string {
	if len(codePart) == 2 {
		return codePart
	}
	if i := strings.IndexByte(codePart, '_'); i >= 0 {
		return codePart[:i]
	}
	return codePart
}

// CountriesFromParts ports anonymous Language.getCountries.
func CountriesFromParts(codePart string) []string {
	if len(codePart) == 2 {
		return []string{""}
	}
	if i := strings.IndexByte(codePart, '_'); i >= 0 {
		return []string{codePart[i+1:]}
	}
	return []string{""}
}
