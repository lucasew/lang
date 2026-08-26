package rules

import (
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AbstractPunctuationCheckRule ports org.languagetool.rules.AbstractPunctuationCheckRule.
//
// Java is abstract with protected abstract isPunctsJoinOk / isPunctuation.
// Go uses function hooks (same pattern as other abstract rule twins); subclasses
// must set both hooks (Init + hooks) before Match — Java always has the abstract
// methods on a concrete subclass.
//
// Java ctor: super(messages); super.setCategory(Categories.PUNCTUATION.getCategory(messages)).
type AbstractPunctuationCheckRule struct {
	Messages map[string]string
	// ID optional override; Java getId always returns "PUNCTUATION_GENERIC_CHECK".
	ID string
	// Description ports getDescription (Java fixed English string).
	Description string
	// Category ports Rule.category (Java PUNCTUATION via ctor).
	Category *Category
	// IsPunctsJoinOk ports protected abstract isPunctsJoinOk(String tokens).
	IsPunctsJoinOk func(tokens string) bool
	// IsPunctuation ports protected abstract isPunctuation(String token).
	IsPunctuation func(token string) bool
}

// InitPunctuationCheckMeta applies Java AbstractPunctuationCheckRule constructor metadata.
func InitPunctuationCheckMeta(r *AbstractPunctuationCheckRule, messages map[string]string) {
	if r == nil {
		return
	}
	r.Messages = messages
	if r.Category == nil {
		r.Category = CatPunctuation.GetCategory(messages)
	}
	if r.Description == "" {
		r.Description = "Use of unusual combination of punctuation characters"
	}
}

// GetID ports AbstractPunctuationCheckRule.getId → "PUNCTUATION_GENERIC_CHECK".
func (r *AbstractPunctuationCheckRule) GetID() string {
	if r != nil && r.ID != "" {
		return r.ID
	}
	return "PUNCTUATION_GENERIC_CHECK"
}

// GetDescription ports AbstractPunctuationCheckRule.getDescription.
func (r *AbstractPunctuationCheckRule) GetDescription() string {
	if r != nil && r.Description != "" {
		return r.Description
	}
	return "Use of unusual combination of punctuation characters"
}

// GetCategory ports Rule.getCategory (Java PUNCTUATION after ctor).
func (r *AbstractPunctuationCheckRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// Match ports AbstractPunctuationCheckRule.match.
//
// Java control flow (tokens *with* whitespace via sentence.getTokens()):
//
//	startTokenIdx = -1; tkns = ""
//	for i, tokenStr:
//	  if isPunctuation(tokenStr):
//	    tkns += tokenStr
//	    if startTokenIdx == -1: startTokenIdx = i
//	    if i < tokens.length-1: continue
//	  if tkns.length() >= 2 && !isPunctsJoinOk(tkns):
//	    RuleMatch from startPos length tkns.length() (UTF-16),
//	    msg "bad duplication...", short "Punctuation problem",
//	    suggestedReplacement tkns.substring(0, 1)
//	  tkns = ""; startTokenIdx = -1
//
// Abstract methods are always present on a Java concrete subclass — hooks must be set.
func (r *AbstractPunctuationCheckRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
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
		// Java: tkns.length() is UTF-16 code units
		if javaUTF16Len(tkns) >= 2 && !r.IsPunctsJoinOk(tkns) {
			msg := "bad duplication or combination of punctuation signs"
			from := tokens[startTokenIdx].GetStartPos()
			// Java: start + tkns.length() (UTF-16 code units)
			to := from + javaUTF16Len(tkns)
			rm := NewRuleMatch(r, sentence, from, to, msg)
			rm.ShortMessage = "Punctuation problem"
			// Java: tkns.substring(0, 1) — first UTF-16 code unit
			rm.SetSuggestedReplacement(javaSubstring01(tkns))
			ruleMatches = append(ruleMatches, rm)
		}
		tkns = ""
		startTokenIdx = -1
	}
	return ruleMatches
}

// javaUTF16Len is Java String.length() (UTF-16 code units).
func javaUTF16Len(s string) int {
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

// javaSubstring01 ports String.substring(0, 1) — first UTF-16 code unit as a Go string.
// For typical BMP punctuation this is one rune; for a supplementary code point Java
// would return a lone high surrogate (invalid Go string), which this rule never sees.
func javaSubstring01(s string) string {
	if s == "" {
		return ""
	}
	u := utf16.Encode([]rune(s))
	if len(u) == 0 {
		return ""
	}
	return string(utf16.Decode(u[:1]))
}
