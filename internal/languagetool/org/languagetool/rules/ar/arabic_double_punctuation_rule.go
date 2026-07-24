package ar

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ArabicDoublePunctuationRule ports org.languagetool.rules.ar.ArabicDoublePunctuationRule.
type ArabicDoublePunctuationRule struct {
	*rules.DoublePunctuationRule
}

func NewArabicDoublePunctuationRule(messages map[string]string) *ArabicDoublePunctuationRule {
	base := rules.NewDoublePunctuationRule(messages)
	base.RuleID = "ARABIC_DOUBLE_PUNCTUATION"
	base.CommaCharacter = "،"
	return &ArabicDoublePunctuationRule{DoublePunctuationRule: base}
}

func (r *ArabicDoublePunctuationRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.DoublePunctuationRule.Match(sentence)
}
