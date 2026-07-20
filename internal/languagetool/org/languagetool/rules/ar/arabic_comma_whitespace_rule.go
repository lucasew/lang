package ar

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ArabicCommaWhitespaceRule ports org.languagetool.rules.ar.ArabicCommaWhitespaceRule.
// Java only overrides getId + getCommaCharacter; Match is CommaWhitespaceRule as-is.
// Glued "هذه،جملة" relies on ArabicWordTokenizer (tokenizing chars include ،).
type ArabicCommaWhitespaceRule struct {
	*rules.CommaWhitespaceRule
}

func NewArabicCommaWhitespaceRule(messages map[string]string) *ArabicCommaWhitespaceRule {
	base := rules.NewCommaWhitespaceRule(messages)
	base.CommaCharacter = "،"
	base.RuleID = "ARABIC_COMMA_PARENTHESIS_WHITESPACE"
	return &ArabicCommaWhitespaceRule{CommaWhitespaceRule: base}
}
