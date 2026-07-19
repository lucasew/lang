package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// LongSentenceRule ports org.languagetool.rules.de.LongSentenceRule.
type LongSentenceRule struct {
	*rules.LongSentenceRule
}

func NewLongSentenceRule(messages map[string]string, maxWords int) *LongSentenceRule {
	if maxWords <= 0 {
		maxWords = 40
	}
	base := rules.NewLongSentenceRule(messages, maxWords)
	base.RuleID = "TOO_LONG_SENTENCE_DE"
	// Java de.LongSentenceRule overrides getDescription / short message.
	base.Description = "Findet lange Sätze"
	base.ShortMsg = "Langer Satz"
	return &LongSentenceRule{LongSentenceRule: base}
}

func (r *LongSentenceRule) GetID() string {
	if r != nil && r.LongSentenceRule != nil {
		return r.LongSentenceRule.GetID()
	}
	return "TOO_LONG_SENTENCE_DE"
}

func (r *LongSentenceRule) GetDescription() string {
	if r != nil && r.LongSentenceRule != nil {
		return r.LongSentenceRule.GetDescription()
	}
	return "Findet lange Sätze"
}

func (r *LongSentenceRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.LongSentenceRule.MatchList(sentences)
}
