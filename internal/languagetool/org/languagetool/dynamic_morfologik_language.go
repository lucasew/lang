package languagetool

import "strings"

// DynamicMorfologikLanguage ports org.languagetool.DynamicMorfologikLanguage.
// Provides Morfologik-backed spelling rule ID/path surfaces for a custom dictionary.
type DynamicMorfologikLanguage struct {
	DynamicLanguage
}

func NewDynamicMorfologikLanguage(name, code, dictPath string) DynamicMorfologikLanguage {
	return DynamicMorfologikLanguage{DynamicLanguage: NewDynamicLanguage(name, code, dictPath)}
}

// SpellerRuleID ports anonymous MorfologikSpellerRule.getId(): code.toUpperCase() + "_SPELLER_RULE".
func (d DynamicMorfologikLanguage) SpellerRuleID() string {
	return strings.ToUpper(d.Code) + "_SPELLER_RULE"
}

// SpellerDictPath ports getFileName() → dictPath.getAbsolutePath().
func (d DynamicMorfologikLanguage) SpellerDictPath() string { return d.DictPath }

// GetFileName ports getFileName (absolute dict path).
func (d DynamicMorfologikLanguage) GetFileName() string { return d.DictPath }

// GetSpellingFileName ports getSpellingFileName → null.
func (d DynamicMorfologikLanguage) GetSpellingFileName() *string { return nil }

// RelevantSpellerRuleIDs ports getRelevantRules singleton list.
func (d DynamicMorfologikLanguage) RelevantSpellerRuleIDs() []string {
	return []string{d.SpellerRuleID()}
}
