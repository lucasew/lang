package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanParagraphRepeatBeginningRule ports org.languagetool.rules.de.GermanParagraphRepeatBeginningRule.
type GermanParagraphRepeatBeginningRule struct {
	*rules.ParagraphRepeatBeginningRule
}

var deArticles = map[string]struct{}{
	"der": {}, "die": {}, "das": {}, "den": {}, "dem": {}, "des": {},
	"ein": {}, "eine": {}, "einer": {}, "einem": {}, "einen": {}, "eines": {},
}

func NewGermanParagraphRepeatBeginningRule(messages map[string]string) *GermanParagraphRepeatBeginningRule {
	base := rules.NewParagraphRepeatBeginningRule(messages)
	base.RuleID = "GERMAN_PARAGRAPH_REPEAT_BEGINNING_RULE"
	base.IsArticle = func(token *languagetool.AnalyzedTokenReadings) bool {
		if token == nil {
			return false
		}
		// Java: hasPosTagStartingWith("ART")
		if token.HasPosTagStartingWith("ART") {
			return true
		}
		_, ok := deArticles[strings.ToLower(token.GetToken())]
		return ok
	}
	return &GermanParagraphRepeatBeginningRule{ParagraphRepeatBeginningRule: base}
}

func (r *GermanParagraphRepeatBeginningRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.ParagraphRepeatBeginningRule.MatchList(sentences)
}
