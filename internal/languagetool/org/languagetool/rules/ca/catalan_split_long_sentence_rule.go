package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// CatalanSplitLongSentenceRule ports the base LongSentenceRule part of
// org.languagetool.rules.ca.CatalanSplitLongSentenceRule (without remote rewrite).
type CatalanSplitLongSentenceRule struct {
	*rules.LongSentenceRule
}

func NewCatalanSplitLongSentenceRule(messages map[string]string, maxWords int) *CatalanSplitLongSentenceRule {
	if maxWords <= 0 {
		maxWords = 40
	}
	base := rules.NewLongSentenceRule(messages, maxWords)
	base.RuleID = "CA_SPLIT_LONG_SENTENCE"
	return &CatalanSplitLongSentenceRule{LongSentenceRule: base}
}

func (r *CatalanSplitLongSentenceRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.LongSentenceRule.MatchList(sentences)
}
