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
		tok := tokens[tokenIdx].GetToken()
		return tok == "\u2014" || tok == "\u2013"
	}
	return &UkrainianCommaWhitespaceRule{CommaWhitespaceRule: base}
}

func (r *UkrainianCommaWhitespaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.CommaWhitespaceRule.Match(sentence)
}
