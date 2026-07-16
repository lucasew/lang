package hunspell

// HunspellNoSuggestionRuleID ports HunspellNoSuggestionRule.RULE_ID.
const HunspellNoSuggestionRuleID = "HUNSPELL_NO_SUGGEST_RULE"

// HunspellNoSuggestionRule ports org.languagetool.rules.spelling.hunspell.HunspellNoSuggestionRule
// as a surface: spell-check without suggestions (dict is pluggable).
type HunspellNoSuggestionRule struct {
	ID   string
	Dict HunspellDictionary
}

func NewHunspellNoSuggestionRule(dict HunspellDictionary) *HunspellNoSuggestionRule {
	return &HunspellNoSuggestionRule{ID: HunspellNoSuggestionRuleID, Dict: dict}
}

func (r *HunspellNoSuggestionRule) GetID() string {
	if r != nil && r.ID != "" {
		return r.ID
	}
	return HunspellNoSuggestionRuleID
}

// IsMisspelled returns true when the dictionary rejects the word.
func (r *HunspellNoSuggestionRule) IsMisspelled(word string) bool {
	if r == nil || r.Dict == nil {
		return false
	}
	return !r.Dict.Spell(word)
}

// Suggest always returns nil (no suggestions by design).
func (r *HunspellNoSuggestionRule) Suggest(word string) []string {
	return nil
}
