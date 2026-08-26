package hunspell

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// HunspellNoSuggestionRuleID ports HunspellNoSuggestionRule.RULE_ID.
const HunspellNoSuggestionRuleID = "HUNSPELL_NO_SUGGEST_RULE"

// HunspellNoSuggestionRule ports org.languagetool.rules.spelling.hunspell.HunspellNoSuggestionRule:
// same match as HunspellRule but never returns suggestions (Java getSuggestions → empty).
type HunspellNoSuggestionRule struct {
	*HunspellRule
}

// NewHunspellNoSuggestionRule builds a no-suggestion hunspell speller for languageCode.
func NewHunspellNoSuggestionRule(languageCode string, dict HunspellDictionary) *HunspellNoSuggestionRule {
	// NewHunspellRule already ApplyDefaultSpellingWordLists (Java SpellingCheckRule.init).
	base := NewHunspellRule(languageCode, dict)
	// Override id to HUNSPELL_NO_SUGGEST_RULE (Java subclass getId).
	if base.SpellingCheckRule != nil {
		base.SpellingCheckRule.ID = HunspellNoSuggestionRuleID
	}
	r := &HunspellNoSuggestionRule{HunspellRule: base}
	// Java overrides getSuggestions → empty; wire SuggestFn so Match/calcSuggestions
	// never call the dictionary (Go embedding is not virtual).
	r.HunspellRule.SuggestFn = func(word string) []string { return nil }
	return r
}

// GetID ports HunspellNoSuggestionRule.getId.
func (r *HunspellNoSuggestionRule) GetID() string {
	return HunspellNoSuggestionRuleID
}

// GetDescription ports messages "desc_spelling_no_suggestions" (en default surface).
func (r *HunspellNoSuggestionRule) GetDescription() string {
	return "Possible spelling mistake (no suggestions)"
}

// Suggest always empty (Java HunspellNoSuggestionRule.getSuggestions).
func (r *HunspellNoSuggestionRule) Suggest(word string) []string {
	return nil
}

// Match ports HunspellRule.match; suggestions stay empty via SuggestFn.
func (r *HunspellNoSuggestionRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if r == nil || r.HunspellRule == nil {
		return nil, nil
	}
	return r.HunspellRule.Match(sentence)
}
