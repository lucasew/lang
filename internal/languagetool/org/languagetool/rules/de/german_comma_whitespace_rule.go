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
	// Java GermanCommaWhitespaceRule.isException: tokenIdx+2 < length, "." then de-Domain.
	base.IsException = func(tokens []*languagetool.AnalyzedTokenReadings, tokenIdx int) bool {
		if tokenIdx+2 < len(tokens) &&
			tokens[tokenIdx].GetToken() == "." &&
			deDomainLabel.MatchString(tokens[tokenIdx+1].GetToken()) {
			return true
		}
		return false
	}
	return &GermanCommaWhitespaceRule{CommaWhitespaceRule: base}
}

// GetURL ports German.java getRelevantRules URL for GermanCommaWhitespaceRule.
func (r *GermanCommaWhitespaceRule) GetURL() string {
	return "https://languagetool.org/insights/de/beitrag/grammatik-leerzeichen/#fehler-1-leerzeichen-vor-und-nach-satzzeichen"
}

func (r *GermanCommaWhitespaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	// Java attaches this (DE rule) so setUrl is visible on matches.
	ms := r.CommaWhitespaceRule.Match(sentence)
	for _, m := range ms {
		if m != nil {
			m.Rule = r
		}
	}
	return ms
}
