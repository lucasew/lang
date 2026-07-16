package br

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// LongSentenceRule wraps the core LongSentenceRule for this language.
type LongSentenceRule struct {
	*rules.LongSentenceRule
}

func NewLongSentenceRule(messages map[string]string, maxWords int) *LongSentenceRule {
	if maxWords <= 0 {
		maxWords = 40
	}
	base := rules.NewLongSentenceRule(messages, maxWords)
	base.RuleID = "TOO_LONG_SENTENCE_BR"
	base.ShortMsg = "Frazenn hir"
	return &LongSentenceRule{LongSentenceRule: base}
}

func (r *LongSentenceRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.LongSentenceRule.MatchList(sentences)
}
