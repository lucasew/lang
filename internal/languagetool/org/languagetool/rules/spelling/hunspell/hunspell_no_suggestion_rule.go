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
	return &HunspellNoSuggestionRule{HunspellRule: base}
}

// GetID ports HunspellNoSuggestionRule.getId.
func (r *HunspellNoSuggestionRule) GetID() string {
	return HunspellNoSuggestionRuleID
}

// Suggest always empty (Java HunspellNoSuggestionRule.getSuggestions).
func (r *HunspellNoSuggestionRule) Suggest(word string) []string {
	return nil
}

// Match ports HunspellRule.match then strips any suggestions.
func (r *HunspellNoSuggestionRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if r == nil || r.HunspellRule == nil {
		return nil, nil
	}
	// Temporarily clear dict suggestions path: parent Match calls Suggest on r.HunspellRule.
	// Call parent Match then strip suggestions on results.
	ms, err := r.HunspellRule.Match(sentence)
	if err != nil {
		return ms, err
	}
	for _, m := range ms {
		if m != nil {
			m.SetSuggestedReplacements(nil)
		}
	}
	return ms, nil
}
