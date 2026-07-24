package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// UkrainianCommaWhitespaceRule ports org.languagetool.rules.uk.UkrainianCommaWhitespaceRule.
type UkrainianCommaWhitespaceRule struct {
	*rules.CommaWhitespaceRule
}

func NewUkrainianCommaWhitespaceRule(messages map[string]string) *UkrainianCommaWhitespaceRule {
	base := rules.NewCommaWhitespaceRule(messages)
	base.IsException = func(tokens []*languagetool.AnalyzedTokenReadings, tokenIdx int) bool {
		if tokenIdx < 0 || tokenIdx >= len(tokens) || tokens[tokenIdx] == nil {
			return false
		}
		t := tokens[tokenIdx].GetToken()
		return t == "\u2014" || t == "\u2013"
	}
	return &UkrainianCommaWhitespaceRule{CommaWhitespaceRule: base}
}
