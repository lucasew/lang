package ar

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// ArabicQuestionMarkWhitespaceRule ports org.languagetool.rules.ar.ArabicQuestionMarkWhitespaceRule.
type ArabicQuestionMarkWhitespaceRule struct {
	*rules.CommaWhitespaceRule
}

func NewArabicQuestionMarkWhitespaceRule(messages map[string]string) *ArabicQuestionMarkWhitespaceRule {
	base := rules.NewCommaWhitespaceRule(messages)
	base.CommaCharacter = "؟"
	base.RuleID = "ARABIC_QM_WHITESPACE"
	return &ArabicQuestionMarkWhitespaceRule{CommaWhitespaceRule: base}
}
