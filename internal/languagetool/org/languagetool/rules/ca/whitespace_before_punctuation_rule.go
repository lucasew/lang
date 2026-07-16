package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// WhitespaceBeforePunctuationRule wraps the core rule for this language.
type WhitespaceBeforePunctuationRule struct {
	*rules.WhitespaceBeforePunctuationRule
}

func NewWhitespaceBeforePunctuationRule(messages map[string]string) *WhitespaceBeforePunctuationRule {
	return &WhitespaceBeforePunctuationRule{WhitespaceBeforePunctuationRule: rules.NewWhitespaceBeforePunctuationRule(messages)}
}

func (r *WhitespaceBeforePunctuationRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.WhitespaceBeforePunctuationRule.Match(sentence)
}
