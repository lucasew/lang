package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanParagraphRepeatBeginningRule ports org.languagetool.rules.de.GermanParagraphRepeatBeginningRule.
type GermanParagraphRepeatBeginningRule struct {
	*rules.ParagraphRepeatBeginningRule
}

func NewGermanParagraphRepeatBeginningRule(messages map[string]string) *GermanParagraphRepeatBeginningRule {
	base := rules.NewParagraphRepeatBeginningRule(messages)
	base.RuleID = "GERMAN_PARAGRAPH_REPEAT_BEGINNING_RULE"
	// Java isArticle: token.hasPosTagStartingWith("ART") only — no surface invent.
	base.IsArticle = func(token *languagetool.AnalyzedTokenReadings) bool {
		return token != nil && token.HasPosTagStartingWith("ART")
	}
	return &GermanParagraphRepeatBeginningRule{ParagraphRepeatBeginningRule: base}
}

func (r *GermanParagraphRepeatBeginningRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.ParagraphRepeatBeginningRule.MatchList(sentences)
}
