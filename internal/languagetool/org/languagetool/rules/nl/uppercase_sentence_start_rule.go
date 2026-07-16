package nl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// UppercaseSentenceStartRule wraps the core UppercaseSentenceStartRule.
type UppercaseSentenceStartRule struct {
	*rules.UppercaseSentenceStartRule
}

func NewUppercaseSentenceStartRule(messages map[string]string) *UppercaseSentenceStartRule {
	return &UppercaseSentenceStartRule{UppercaseSentenceStartRule: rules.NewUppercaseSentenceStartRule(messages, "nl")}
}

func (r *UppercaseSentenceStartRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.UppercaseSentenceStartRule.MatchList(sentences)
}
