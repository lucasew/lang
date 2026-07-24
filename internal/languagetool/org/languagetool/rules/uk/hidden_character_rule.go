package uk

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

const hiddenChar = '\u00AD' // soft hyphen

// HiddenCharacterRule ports org.languagetool.rules.uk.HiddenCharacterRule.
// Java: setCategory(Categories.MISC).
type HiddenCharacterRule struct {
	Messages map[string]string
	Category *rules.Category
}

func NewHiddenCharacterRule(messages map[string]string) *HiddenCharacterRule {
	return &HiddenCharacterRule{
		Messages: messages,
		Category: rules.CatMisc.GetCategory(messages),
	}
}

func (r *HiddenCharacterRule) GetID() string { return "UK_HIDDEN_CHARS" }

func (r *HiddenCharacterRule) GetDescription() string {
	return "Приховані символи: знак м’якого перенесення"
}

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

func (r *HiddenCharacterRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var ruleMatches []*rules.RuleMatch
	for _, tokenReadings := range sentence.GetTokensWithoutWhitespace() {
		tokenString := tokenReadings.GetToken()
		if strings.ContainsRune(tokenString, hiddenChar) {
			ruleMatches = append(ruleMatches, r.createRuleMatch(tokenReadings, sentence))
		}
	}
	return ruleMatches
}

func (r *HiddenCharacterRule) createRuleMatch(readings *languagetool.AnalyzedTokenReadings, sentence *languagetool.AnalyzedSentence) *rules.RuleMatch {
	tokenString := readings.GetToken()
	replacement := strings.ReplaceAll(tokenString, string(hiddenChar), "")
	highlighted := strings.ReplaceAll(tokenString, string(hiddenChar), "-")
	msg := tokenString + " містить невидимий знак м’якого перенесення: «" + highlighted + "», виправлення: " + replacement
	rm := rules.NewRuleMatch(r, sentence, readings.GetStartPos(), readings.GetEndPos(), msg)
	rm.ShortMessage = "Приховані символи"
	rm.SetSuggestedReplacement(replacement)
	return rm
}
