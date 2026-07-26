package uk

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// HIDDEN_CHAR ports Java HiddenCharacterRule.HIDDEN_CHAR (U+00AD soft hyphen).
const HIDDEN_CHAR = '\u00AD'

// HiddenCharacterRule ports org.languagetool.rules.uk.HiddenCharacterRule.
// Java: setCategory(Categories.MISC.getCategory(messages)).
type HiddenCharacterRule struct {
	Messages map[string]string
	Category *rules.Category
}

// NewHiddenCharacterRule ports HiddenCharacterRule(ResourceBundle messages).
func NewHiddenCharacterRule(messages map[string]string) *HiddenCharacterRule {
	return &HiddenCharacterRule{
		Messages: messages,
		Category: rules.CatMisc.GetCategory(messages),
	}
}

// GetID ports HiddenCharacterRule.getId → "UK_HIDDEN_CHARS".
func (r *HiddenCharacterRule) GetID() string { return "UK_HIDDEN_CHARS" }

// GetDescription ports HiddenCharacterRule.getDescription.
func (r *HiddenCharacterRule) GetDescription() string {
	return "Приховані символи: знак м’якого перенесення"
}

// GetShort ports HiddenCharacterRule.getShort.
func (r *HiddenCharacterRule) GetShort() string {
	return "Приховані символи"
}

// GetCategory ports Rule.getCategory (Java MISC).
func (r *HiddenCharacterRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// GetSuggestion ports HiddenCharacterRule.getSuggestion(String word):
// soft hyphens highlighted as '-' inside «…».
func (r *HiddenCharacterRule) GetSuggestion(word string) string {
	highlighted := strings.ReplaceAll(word, string(HIDDEN_CHAR), "-")
	return " містить невидимий знак м’якого перенесення: «" + highlighted + "», виправлення: "
}

// Match ports HiddenCharacterRule.match.
func (r *HiddenCharacterRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var ruleMatches []*rules.RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	for _, tokenReadings := range tokens {
		tokenString := tokenReadings.GetToken()
		// Java: tokenString.indexOf(HIDDEN_CHAR) != -1
		if strings.ContainsRune(tokenString, HIDDEN_CHAR) {
			ruleMatches = append(ruleMatches, r.createRuleMatch(tokenReadings, sentence))
		}
	}
	return ruleMatches
}

// createRuleMatch ports HiddenCharacterRule.createRuleMatch.
func (r *HiddenCharacterRule) createRuleMatch(readings *languagetool.AnalyzedTokenReadings, sentence *languagetool.AnalyzedSentence) *rules.RuleMatch {
	tokenString := readings.GetToken()
	// Java: tokenString.replace(HIDDEN_CHAR.toString(), "")
	replacement := strings.ReplaceAll(tokenString, string(HIDDEN_CHAR), "")
	// Java: msg = tokenString + getSuggestion(tokenString) + replacement
	msg := tokenString + r.GetSuggestion(tokenString) + replacement
	// Java: new RuleMatch(this, sentence, start, end, msg, getShort())
	rm := rules.NewRuleMatch(r, sentence, readings.GetStartPos(), readings.GetEndPos(), msg)
	rm.ShortMessage = r.GetShort()
	// Java: setSuggestedReplacements(Arrays.asList(replacement))
	rm.SetSuggestedReplacements([]string{replacement})
	return rm
}
