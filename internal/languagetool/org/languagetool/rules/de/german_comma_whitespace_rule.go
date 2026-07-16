package de

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

var deDomainLabel = regexp.MustCompile(`^[a-z]{2,10}-Domains?$`)

// GermanCommaWhitespaceRule ports org.languagetool.rules.de.GermanCommaWhitespaceRule.
type GermanCommaWhitespaceRule struct {
	*rules.CommaWhitespaceRule
}

func NewGermanCommaWhitespaceRule(messages map[string]string) *GermanCommaWhitespaceRule {
	base := rules.NewCommaWhitespaceRule(messages)
	base.IsException = func(tokens []*languagetool.AnalyzedTokenReadings, tokenIdx int) bool {
		if tokenIdx+1 < len(tokens) &&
			tokens[tokenIdx].GetToken() == "." &&
			deDomainLabel.MatchString(tokens[tokenIdx+1].GetToken()) {
			return true
		}
		return false
	}
	return &GermanCommaWhitespaceRule{CommaWhitespaceRule: base}
}

func (r *GermanCommaWhitespaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.CommaWhitespaceRule.Match(sentence)
}
