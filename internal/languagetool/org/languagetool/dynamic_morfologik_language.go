package languagetool

import (
	"strings"
)

// DynamicMorfologikLanguage ports org.languagetool.DynamicMorfologikLanguage.
// Provides a Morfologik-backed spelling rule ID surface for a custom dictionary path.
type DynamicMorfologikLanguage struct {
	DynamicLanguage
}

func NewDynamicMorfologikLanguage(name, code, dictPath string) DynamicMorfologikLanguage {
	return DynamicMorfologikLanguage{DynamicLanguage: NewDynamicLanguage(name, code, dictPath)}
}

// SpellerRuleID ports the anonymous MorfologikSpellerRule.getId() (CODE_SPELLER_RULE).
func (d DynamicMorfologikLanguage) SpellerRuleID() string {
	return strings.ToUpper(d.Code) + "_SPELLER_RULE"
}

// SpellerDictPath is the absolute dict path used by the dynamic speller rule.
func (d DynamicMorfologikLanguage) SpellerDictPath() string { return d.DictPath }

// RelevantSpellerRuleIDs returns the single dynamic speller rule id (Java getRelevantRules).
func (d DynamicMorfologikLanguage) RelevantSpellerRuleIDs() []string {
	return []string{d.SpellerRuleID()}
}
