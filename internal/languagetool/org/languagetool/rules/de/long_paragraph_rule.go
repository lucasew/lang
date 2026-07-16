package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// LongParagraphRule is a thin wrapper around the core LongParagraphRule for German configs.
type LongParagraphRule struct {
	*rules.LongParagraphRule
}

func NewLongParagraphRule(messages map[string]string, maxWords int) *LongParagraphRule {
	if maxWords <= 0 {
		maxWords = 150
	}
	base := rules.NewLongParagraphRule(messages, maxWords)
	return &LongParagraphRule{LongParagraphRule: base}
}

func (r *LongParagraphRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.LongParagraphRule.MatchList(sentences)
}
