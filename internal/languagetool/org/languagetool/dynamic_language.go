package languagetool

import (
	"path/filepath"
	"strings"
)

// DynamicLanguage ports org.languagetool.DynamicLanguage (package-private abstract in Java).
// Metadata surface for dict-backed spellcheck-only languages without full Language hierarchy.
type DynamicLanguage struct {
	Name     string
	Code     string // may include variant, e.g. en-US
	DictPath string // Java: File dictPath
}

// NewDynamicLanguage ports DynamicLanguage(String name, String code, File dictPath).
// Java uses Objects.requireNonNull — empty strings are allowed; Go cannot express null File/String.
func NewDynamicLanguage(name, code, dictPath string) DynamicLanguage {
	return DynamicLanguage{Name: name, Code: code, DictPath: dictPath}
}

// GetShortCode ports getShortCode — DASH.matcher(code).replaceFirst("") with Pattern "-.*".
func (d DynamicLanguage) GetShortCode() string {
	if i := strings.IndexByte(d.Code, '-'); i >= 0 {
		return d.Code[:i]
	}
	return d.Code
}

func (d DynamicLanguage) GetName() string { return d.Name }

// GetShortCodeWithCountryAndVariant returns the full code field (Language default uses code).
func (d DynamicLanguage) GetShortCodeWithCountryAndVariant() string { return d.Code }

// GetRuleFileNames ports getRuleFileNames → empty list.
func (d DynamicLanguage) GetRuleFileNames() []string { return []string{} }

// GetPatternRules ports getPatternRules → empty list (typed as any until pattern rules twin).
func (d DynamicLanguage) GetPatternRules() []any { return []any{} }

// GetCommonWordsPath ports getCommonWordsPath:
// new File(dictPath.getParentFile(), "common_words.txt").getAbsolutePath()
func (d DynamicLanguage) GetCommonWordsPath() string {
	parent := filepath.Dir(d.DictPath)
	return filepath.Join(parent, "common_words.txt")
}

// GetCountries ports getCountries → empty array.
func (d DynamicLanguage) GetCountries() []string { return []string{} }

// GetMaintainers ports getMaintainers → empty Contributor array (typed as any).
func (d DynamicLanguage) GetMaintainers() []any { return []any{} }

// IsSpellcheckOnlyLanguage ports isSpellcheckOnlyLanguage → true.
func (d DynamicLanguage) IsSpellcheckOnlyLanguage() bool { return true }
