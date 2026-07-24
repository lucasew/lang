package ar

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// ArabicSemiColonWhitespaceRule ports org.languagetool.rules.ar.ArabicSemiColonWhitespaceRule.
type ArabicSemiColonWhitespaceRule struct {
	*rules.CommaWhitespaceRule
}

func NewArabicSemiColonWhitespaceRule(messages map[string]string) *ArabicSemiColonWhitespaceRule {
	base := rules.NewCommaWhitespaceRule(messages)
	base.CommaCharacter = "؛"
	base.RuleID = "ARABIC_SC_WHITESPACE"
	return &ArabicSemiColonWhitespaceRule{CommaWhitespaceRule: base}
}
