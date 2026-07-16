package fa

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PersianCommaWhitespaceRule ports org.languagetool.rules.fa.PersianCommaWhitespaceRule.
type PersianCommaWhitespaceRule struct {
	*rules.CommaWhitespaceRule
}

func NewPersianCommaWhitespaceRule(messages map[string]string) *PersianCommaWhitespaceRule {
	base := rules.NewCommaWhitespaceRule(messages)
	base.RuleID = "PERSIAN_COMMA_PARENTHESIS_WHITESPACE"
	base.CommaCharacter = "،"
	return &PersianCommaWhitespaceRule{CommaWhitespaceRule: base}
}

func (r *PersianCommaWhitespaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.CommaWhitespaceRule.Match(sentence)
}
