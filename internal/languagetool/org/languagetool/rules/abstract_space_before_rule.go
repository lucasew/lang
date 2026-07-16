package rules

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AbstractSpaceBeforeRule ports org.languagetool.rules.AbstractSpaceBeforeRule.
type AbstractSpaceBeforeRule struct {
	Messages     map[string]string
	ID           string
	Description  string
	ShortMsg     string
	Suggestion   string
	// Conjunctions matches tokens that require a space before them.
	Conjunctions *regexp.Regexp
}

func (r *AbstractSpaceBeforeRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "SPACE_BEFORE_CONJUNCTION"
}

func (r *AbstractSpaceBeforeRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	tokens := sentence.GetTokens()
	msg := r.Suggestion
	if msg == "" {
		msg = "Missing white space before conjunction"
	}
	short := r.ShortMsg
	if short == "" {
		short = "Missing white space"
	}
	for i := 1; i < len(tokens); i++ {
		token := tokens[i].GetToken()
		if r.Conjunctions == nil || !r.Conjunctions.MatchString(token) {
			continue
		}
		prev := tokens[i-1].GetToken()
		if prev == " " || prev == "(" {
			continue
		}
		// SENT_START empty token: still flag (Java does)
		from := tokens[i].GetStartPos()
		to := from + utf16Len(token)
		rm := NewRuleMatch(r, sentence, from, to, msg)
		rm.ShortMessage = short
		rm.SetSuggestedReplacement(" " + token)
		ruleMatches = append(ruleMatches, rm)
	}
	return ruleMatches
}
