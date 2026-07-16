package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// LongSentenceRule ports org.languagetool.rules.en.LongSentenceRule.
type LongSentenceRule struct {
	*rules.LongSentenceRule
}

func NewLongSentenceRule(messages map[string]string, maxWords int) *LongSentenceRule {
	if maxWords <= 0 {
		maxWords = 40
	}
	base := rules.NewLongSentenceRule(messages, maxWords)
	base.ShortMsg = "Long sentence"
	return &LongSentenceRule{LongSentenceRule: base}
}

func (r *LongSentenceRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.LongSentenceRule.MatchList(sentences)
}
