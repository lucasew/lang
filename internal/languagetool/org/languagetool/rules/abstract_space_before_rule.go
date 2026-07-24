package rules

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AbstractSpaceBeforeRule ports org.languagetool.rules.AbstractSpaceBeforeRule.
// Java ctor: super.setCategory(Categories.MISC.getCategory(messages)).
// Language parameter is accepted by Java but unused in the abstract class.
type AbstractSpaceBeforeRule struct {
	Messages map[string]string
	// ID overrides getId(); empty → "SPACE_BEFORE_CONJUNCTION" (Java base).
	ID string
	// Description overrides getDescription(); empty → Java fixed English string.
	Description string
	// ShortMsg ports protected getShort(); empty → Java default.
	ShortMsg string
	// Suggestion ports protected getSuggestion(); empty → Java default.
	Suggestion string
	// Conjunctions ports protected abstract getConjunctions() Pattern.
	// Match uses Java Matcher.matches() (full-string) semantics on each token.
	Conjunctions *regexp.Regexp
	// Category ports Rule.category (Java MISC via ctor).
	Category *Category
	// DefaultOff ports Rule.setDefaultOff (e.g. PersianSpaceBeforeRule ctor).
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

// GetID ports AbstractSpaceBeforeRule.getId.
func (r *AbstractSpaceBeforeRule) GetID() string {
	if r != nil && r.ID != "" {
		return r.ID
	}
	return "SPACE_BEFORE_CONJUNCTION"
}

// IsDefaultOff ports Rule.isDefaultOff.
func (r *AbstractSpaceBeforeRule) IsDefaultOff() bool {
	return r != nil && r.DefaultOff
}

// GetDescription ports AbstractSpaceBeforeRule.getDescription.
func (r *AbstractSpaceBeforeRule) GetDescription() string {
	if r != nil && r.Description != "" {
		return r.Description
	}
	// Java: return "Checks for missing space before some conjunctions";
	return "Checks for missing space before some conjunctions"
}

// GetCategory ports Rule.getCategory.
func (r *AbstractSpaceBeforeRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// getShort ports protected AbstractSpaceBeforeRule.getShort.
func (r *AbstractSpaceBeforeRule) getShort() string {
	if r != nil && r.ShortMsg != "" {
		return r.ShortMsg
	}
	return "Missing white space"
}

// getSuggestion ports protected AbstractSpaceBeforeRule.getSuggestion.
func (r *AbstractSpaceBeforeRule) getSuggestion() string {
	if r != nil && r.Suggestion != "" {
		return r.Suggestion
	}
	return "Missing white space before conjunction"
}

// getConjunctions ports protected abstract getConjunctions().
func (r *AbstractSpaceBeforeRule) getConjunctions() *regexp.Regexp {
	if r == nil {
		return nil
	}
	return r.Conjunctions
}

// Match ports AbstractSpaceBeforeRule.match (final in Java).
//
// Java control flow:
//
//	tokens = sentence.getTokens()
//	for i = 1; i < tokens.length; i++
//	  token = tokens[i].getToken()
//	  if getConjunctions().matcher(token).matches()  // full-string
//	    previousToken = tokens[i-1].getToken()
//	    if !(previousToken.equals(" ") || previousToken.equals("("))
//	      match at startPos .. startPos+token.length()  // Java String.length = UTF-16
//	      message=getSuggestion(), shortMessage=getShort()
//	      suggestedReplacement = " " + token
func (r *AbstractSpaceBeforeRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	tokens := sentence.GetTokens()
	msg := r.getSuggestion()
	short := r.getShort()
	conj := r.getConjunctions()
	for i := 1; i < len(tokens); i++ {
		token := tokens[i].GetToken()
		// Java: Matcher.matches() — entire token region, not find()/substring.
		if !javaMatcherMatches(conj, token) {
			continue
		}
		previousToken := tokens[i-1].GetToken()
		if previousToken == " " || previousToken == "(" {
			continue
		}
		// SENT_START empty token "": neither " " nor "(" → still flag (Java does).
		pos := tokens[i].GetStartPos()
		// Java: pos + token.length() — UTF-16 code units.
		to := pos + utf16Len(token)
		potentialRuleMatch := NewRuleMatch(r, sentence, pos, to, msg)
		potentialRuleMatch.ShortMessage = short
		potentialRuleMatch.SetSuggestedReplacement(" " + token)
		ruleMatches = append(ruleMatches, potentialRuleMatch)
	}
	return ruleMatches
}

// javaMatcherMatches ports Java Matcher.matches(): the entire input must match the pattern.
// Go regexp.MatchString is unanchored (substring); this is full-region like Java.
func javaMatcherMatches(re *regexp.Regexp, s string) bool {
	if re == nil {
		return false
	}
	loc := re.FindStringIndex(s)
	return loc != nil && loc[0] == 0 && loc[1] == len(s)
}
