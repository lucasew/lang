package nl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// LongSentenceRule wraps the core LongSentenceRule for this language.
type LongSentenceRule struct {
	*rules.LongSentenceRule
}

func NewLongSentenceRule(messages map[string]string, maxWords int) *LongSentenceRule {
	// Java Dutch.getRelevantRules: new LongSentenceRule(messages, userConfig, 40)
	// ID stays LongSentenceRule.RULE_ID = TOO_LONG_SENTENCE (not a language-suffixed invent).
	if maxWords <= 0 {
		maxWords = 40
	}
	base := rules.NewLongSentenceRule(messages, maxWords)
	base.ShortMsg = "Lange zin"
	return &LongSentenceRule{LongSentenceRule: base}
}

func (r *LongSentenceRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.LongSentenceRule.MatchList(sentences)
}
