package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AbstractPunctuationCheckRule ports org.languagetool.rules.AbstractPunctuationCheckRule.
type AbstractPunctuationCheckRule struct {
	Messages map[string]string
	ID       string
	// IsPunctsJoinOk reports whether a run of punctuation tokens is allowed.
	IsPunctsJoinOk func(tokens string) bool
	// IsPunctuation reports whether a single token is punctuation for this rule.
	IsPunctuation func(token string) bool
}

func (r *AbstractPunctuationCheckRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "PUNCTUATION_GENERIC_CHECK"
}

// Match ports AbstractPunctuationCheckRule.match (uses tokens *with* whitespace).
func (r *AbstractPunctuationCheckRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	if r.IsPunctsJoinOk == nil || r.IsPunctuation == nil {
		return nil
	}
	var ruleMatches []*RuleMatch
	tokens := sentence.GetTokens()
	startTokenIdx := -1
	tkns := ""
	for i := 0; i < len(tokens); i++ {
		tokenStr := tokens[i].GetToken()
		if r.IsPunctuation(tokenStr) {
			tkns += tokenStr
			if startTokenIdx == -1 {
				startTokenIdx = i
			}
			if i < len(tokens)-1 {
				continue
			}
		}
		if len(tkns) >= 2 && !r.IsPunctsJoinOk(tkns) {
			msg := "bad duplication or combination of punctuation signs"
			from := tokens[startTokenIdx].GetStartPos()
			// Java: start + tkns.length() as UTF-16 code units for BMP punctuation
			to := from + punctUTF16Len(tkns)
			rm := NewRuleMatch(r, sentence, from, to, msg)
			rm.ShortMessage = "Punctuation problem"
			rm.SetSuggestedReplacement(string([]rune(tkns)[0]))
			ruleMatches = append(ruleMatches, rm)
		}
		tkns = ""
		startTokenIdx = -1
	}
	return ruleMatches
}

func punctUTF16Len(s string) int {
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
