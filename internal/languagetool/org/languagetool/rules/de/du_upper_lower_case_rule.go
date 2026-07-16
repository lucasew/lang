package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// DuUpperLowerCaseRule ports org.languagetool.rules.de.DuUpperLowerCaseRule.
type DuUpperLowerCaseRule struct {
	Messages map[string]string
}

var duLowerWords = map[string]struct{}{
	"du": {}, "dir": {}, "dich": {}, "dein": {}, "deine": {}, "deines": {}, "deins": {},
	"deiner": {}, "deinen": {}, "deinem": {},
	"euch": {}, "euer": {}, "eure": {}, "euere": {}, "euren": {}, "eueren": {}, "euern": {},
	"eurer": {}, "euerer": {}, "eurem": {}, "euerem": {}, "eures": {}, "eueres": {},
}

var duSkipPrev = map[string]struct{}{
	"\"": {}, "„": {}, "‚": {}, ":": {}, "»": {}, "«": {}, "“": {}, "-": {}, "–": {},
	"*": {}, "•": {}, "\u2063": {}, "\u25E6": {}, "\u00B7": {},
}

func NewDuUpperLowerCaseRule(messages map[string]string) *DuUpperLowerCaseRule {
	return &DuUpperLowerCaseRule{Messages: messages}
}

func (r *DuUpperLowerCaseRule) GetID() string { return "DE_DU_UPPER_LOWER" }

func (r *DuUpperLowerCaseRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var firstUse string
	var ruleMatches []*rules.RuleMatch
	pos := 0
	for _, sentence := range sentences {
		tokens := sentence.GetTokensWithoutWhitespace()
		for i := 0; i < len(tokens); i++ {
			// Skip after sentence start or dialogue punctuation (Java: i>0 && …).
			if i > 0 {
				prev := tokens[i-1]
				if prev.IsSentenceStart() {
					continue
				}
				if _, ok := duSkipPrev[prev.GetToken()]; ok {
					continue
				}
			}
			word := tokens[i].GetToken()
			lcWord := strings.ToLower(word)
			if _, ok := duLowerWords[lcWord]; !ok {
				continue
			}
			if firstUse == "" {
				firstUse = word
				continue
			}
			firstUseIsUpper := tools.StartsWithUppercase(firstUse)
			var msg, replacement string
			if firstUseIsUpper && !tools.StartsWithUppercase(word) {
				replacement = tools.UppercaseFirstChar(word)
				msg = "Vorher wurde bereits '" + firstUse + "' großgeschrieben. " +
					"Aus Gründen der Einheitlichkeit '" + replacement + "' hier auch großschreiben?"
			} else if !firstUseIsUpper && tools.StartsWithUppercase(word) && !tools.IsAllUppercase(word) {
				replacement = tools.LowercaseFirstChar(word)
				msg = "Vorher wurde bereits '" + firstUse + "' kleingeschrieben. " +
					"Aus Gründen der Einheitlichkeit '" + replacement + "' hier auch kleinschreiben?"
			}
			if msg != "" {
				rm := rules.NewRuleMatch(r, sentence, pos+tokens[i].GetStartPos(), pos+tokens[i].GetEndPos(), msg)
				rm.SetSuggestedReplacement(replacement)
				ruleMatches = append(ruleMatches, rm)
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}
