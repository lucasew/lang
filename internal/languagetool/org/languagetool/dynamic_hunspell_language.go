package languagetool

import (
	"regexp"
	"strings"
)

// DynamicHunspellLanguage ports org.languagetool.DynamicHunspellLanguage.
// Speller rule construction (HunspellRule anonymous subclass) is partial until
// full Language + HunspellRule stack is wired; ID/path surfaces match Java.
type DynamicHunspellLanguage struct {
	DynamicLanguage
}

func NewDynamicHunspellLanguage(name, code, dictPath string) DynamicHunspellLanguage {
	return DynamicHunspellLanguage{DynamicLanguage: NewDynamicLanguage(name, code, dictPath)}
}

// SpellerRuleID ports anonymous HunspellRule.getId(): code.toUpperCase() + "_SPELLER_RULE".
func (d DynamicHunspellLanguage) SpellerRuleID() string {
	return strings.ToUpper(d.Code) + "_SPELLER_RULE"
}

// dicSuffixRE mirrors Java replaceAll(".dic$", "") — regex, not literal.
var dicSuffixRE = regexp.MustCompile(`.dic$`)

// DictFilenameInResources ports getDictFilenameInResources:
// dictPath.getAbsolutePath().replaceAll(".dic$", "")
func (d DynamicHunspellLanguage) DictFilenameInResources() string {
	return dicSuffixRE.ReplaceAllString(d.DictPath, "")
}

// GetSpellingFileName ports getSpellingFileName → null.
func (d DynamicHunspellLanguage) GetSpellingFileName() *string { return nil }

// RelevantSpellerRuleIDs ports getRelevantRules returning singleton list of the dynamic rule.
func (d DynamicHunspellLanguage) RelevantSpellerRuleIDs() []string {
	return []string{d.SpellerRuleID()}
}
