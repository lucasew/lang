package fa

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PersianDoublePunctuationRule ports org.languagetool.rules.fa.PersianDoublePunctuationRule.
type PersianDoublePunctuationRule struct {
	*rules.DoublePunctuationRule
}

func NewPersianDoublePunctuationRule(messages map[string]string) *PersianDoublePunctuationRule {
	base := rules.NewDoublePunctuationRule(messages)
	base.RuleID = "PERSIAN_DOUBLE_PUNCTUATION"
	base.CommaCharacter = "،"
	return &PersianDoublePunctuationRule{DoublePunctuationRule: base}
}

func (r *PersianDoublePunctuationRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.DoublePunctuationRule.Match(sentence)
}
