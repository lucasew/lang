package rules

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AbstractSpaceBeforeRule ports org.languagetool.rules.AbstractSpaceBeforeRule.
// Java ctor: setCategory(Categories.MISC.getCategory(messages)).
type AbstractSpaceBeforeRule struct {
	Messages    map[string]string
	ID          string
	Description string
	ShortMsg    string
	Suggestion  string
	// Conjunctions matches tokens that require a space before them.
	Conjunctions *regexp.Regexp
	// Category ports Rule.category (Java MISC).
	Category *Category
	// DefaultOff ports Rule.setDefaultOff (e.g. PersianSpaceBeforeRule).
	DefaultOff bool
}

// InitSpaceBeforeMeta applies Java AbstractSpaceBeforeRule constructor metadata.
func InitSpaceBeforeMeta(r *AbstractSpaceBeforeRule, messages map[string]string) {
	if r == nil {
		return
	}
	r.Messages = messages
	if r.Category == nil {
		r.Category = CatMisc.GetCategory(messages)
	}
}

func (r *AbstractSpaceBeforeRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "SPACE_BEFORE_CONJUNCTION"
}

// IsDefaultOff ports Rule.isDefaultOff.
func (r *AbstractSpaceBeforeRule) IsDefaultOff() bool {
	return r != nil && r.DefaultOff
}

func (r *AbstractSpaceBeforeRule) GetDescription() string {
	if r != nil && r.Description != "" {
		return r.Description
	}
	// Java AbstractSpaceBeforeRule.getDescription
	return "Checks for missing space before some conjunctions"
}

// GetCategory ports Rule.getCategory.
func (r *AbstractSpaceBeforeRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
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
