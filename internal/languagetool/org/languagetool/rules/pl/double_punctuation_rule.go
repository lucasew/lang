package pl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DoublePunctuationRule wraps the core DoublePunctuationRule for this language.
type DoublePunctuationRule struct {
	*rules.DoublePunctuationRule
}

func NewDoublePunctuationRule(messages map[string]string) *DoublePunctuationRule {
	return &DoublePunctuationRule{DoublePunctuationRule: rules.NewDoublePunctuationRule(messages)}
}

func (r *DoublePunctuationRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.DoublePunctuationRule.Match(sentence)
}
