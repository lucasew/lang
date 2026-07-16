package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// DoublePunctuationRule ports org.languagetool.rules.DoublePunctuationRule.
type DoublePunctuationRule struct {
	Messages       map[string]string
	RuleID         string // override GetID when set (e.g. DE_DOUBLE_PUNCTUATION)
	DotMessage     string // override two-dots message when set
	CommaCharacter string // override comma character (Arabic/Persian "،")
}

func NewDoublePunctuationRule(messages map[string]string) *DoublePunctuationRule {
	return &DoublePunctuationRule{Messages: messages}
}

func (r *DoublePunctuationRule) GetID() string {
	if r.RuleID != "" {
		return r.RuleID
	}
	return "DOUBLE_PUNCTUATION"
}

func (r *DoublePunctuationRule) GetCommaCharacter() string {
	if r.CommaCharacter != "" {
		return r.CommaCharacter
	}
	return ","
}

func (r *DoublePunctuationRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	startPos := 0
	dotCount, commaCount := 0, 0
	commaChar := r.GetCommaCharacter()
	for i := 1; i < len(tokens); i++ {
		token := tokens[i].GetToken()
		var nextToken, prevPrevToken string
		if i < len(tokens)-1 {
			nextToken = tokens[i+1].GetToken()
		}
		if i > 1 {
			prevPrevToken = tokens[i-2].GetToken()
		}
		if token == "." {
			dotCount++
			commaCount = 0
			startPos = tokens[i].GetStartPos()
		} else if token == commaChar {
			commaCount++
			dotCount = 0
			startPos = tokens[i].GetStartPos()
		}

		if dotCount == 2 && nextToken != "." && nextToken != "…" &&
			token != "/" && nextToken != "/" &&
			token != "\\" && nextToken != "\\" &&
			prevPrevToken != "?" && prevPrevToken != "!" &&
			prevPrevToken != "…" && prevPrevToken != "." {
			fromPos := startPos - 1
			if fromPos < 0 {
				fromPos = 0
			}
			msg := r.DotMessage
			if msg == "" {
				msg = r.Messages["two_dots"]
			}
			if msg == "" {
				msg = "Two consecutive dots"
			}
			rm := NewRuleMatch(r, sentence, fromPos, startPos+1, msg)
			rm.SuggestedReplacements = []string{".", "…"}
			ruleMatches = append(ruleMatches, rm)
			dotCount = 0
		} else if commaCount == 2 && nextToken != commaChar {
			fromPos := startPos - 1
			if fromPos < 0 {
				fromPos = 0
			}
			msg := r.Messages["two_commas"]
			if msg == "" {
				msg = "Two consecutive commas"
			}
			rm := NewRuleMatch(r, sentence, fromPos, startPos+1, msg)
			rm.SetSuggestedReplacement(commaChar)
			ruleMatches = append(ruleMatches, rm)
			commaCount = 0
		}
		if token != "." && token != commaChar {
			dotCount = 0
			commaCount = 0
		}
	}
	return ruleMatches
}
