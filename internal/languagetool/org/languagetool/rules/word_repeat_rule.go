package rules

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// WordRepeatRule ports org.languagetool.rules.WordRepeatRule.
type WordRepeatRule struct {
	Messages map[string]string
	// ExtraIgnore is called from Ignore for language-specific exceptions (e.g. EnglishWordRepeatRule).
	ExtraIgnore func(tokens []*languagetool.AnalyzedTokenReadings, position int) bool
	// CreateMatchFn optional override for createRuleMatch (e.g. Ukrainian І/і suggestion).
	CreateMatchFn func(r *WordRepeatRule, sentence *languagetool.AnalyzedSentence, prevToken, token string, prevPos, pos int, msg string) *RuleMatch
	// IDOverride when non-empty replaces the default WORD_REPEAT_RULE id.
	IDOverride string
}

func NewWordRepeatRule(messages map[string]string) *WordRepeatRule {
	return &WordRepeatRule{Messages: messages}
}

func (r *WordRepeatRule) GetID() string {
	if r.IDOverride != "" {
		return r.IDOverride
	}
	return "WORD_REPEAT_RULE"
}

func (r *WordRepeatRule) Ignore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if r.ExtraIgnore != nil && r.ExtraIgnore(tokens, position) {
		return true
	}
	for _, w := range []string{"Phi", "Li", "Xiao", "Duran", "Wagga", "Abdullah", "Nwe", "Pago", "Cao"} {
		if r.wordRepetitionOf(w, tokens, position) {
			return true
		}
	}
	return false
}

func (r *WordRepeatRule) wordRepetitionOf(word string, tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	return position > 0 &&
		tokens[position-1].GetToken() == word &&
		tokens[position].GetToken() == word
}

func (r *WordRepeatRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	prevToken := ""
	msg := r.Messages["repetition"]
	if msg == "" {
		msg = "Word repetition"
	}
	for i := 1; i < len(tokens); i++ {
		token := tokens[i].GetToken()
		if tokens[i].IsImmunized() {
			prevToken = ""
			continue
		}
		if isWord(token) && strings.EqualFold(prevToken, token) && !r.Ignore(tokens, i) {
			prevPos := tokens[i-1].GetStartPos()
			pos := tokens[i].GetStartPos()
			var rm *RuleMatch
			if r.CreateMatchFn != nil {
				rm = r.CreateMatchFn(r, sentence, prevToken, token, prevPos, pos, msg)
			} else {
				rm = NewRuleMatch(r, sentence, prevPos, pos+utf16Len(prevToken), msg)
				rm.SetSuggestedReplacement(prevToken)
			}
			ruleMatches = append(ruleMatches, rm)
		}
		prevToken = token
	}
	return ruleMatches
}

func isWord(token string) bool {
	if tools.IsEmoji(token) {
		return false
	}
	if tools.IsNumericSpace(token) {
		return false
	}
	runes := []rune(token)
	if len(runes) == 1 {
		if !unicode.IsLetter(runes[0]) {
			return false
		}
	}
	return true
}

func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	return n
}
