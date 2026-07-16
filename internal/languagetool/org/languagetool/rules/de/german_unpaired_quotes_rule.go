package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanUnpairedQuotesRule ports org.languagetool.rules.de.GermanUnpairedQuotesRule.
type GermanUnpairedQuotesRule struct {
	*rules.GenericUnpairedQuotesRule
}

func NewGermanUnpairedQuotesRule(messages map[string]string) *GermanUnpairedQuotesRule {
	start := []string{"„", "»", "«", "\"", "'", "‚", "›", "‹"}
	end := []string{"“", "«", "»", "\"", "'", "‘", "‹", "›"}
	base := rules.NewGenericUnpairedQuotesRule(messages, start, end)
	base.SetRuleID("DE_UNPAIRED_QUOTES")
	return &GermanUnpairedQuotesRule{GenericUnpairedQuotesRule: base}
}

func (r *GermanUnpairedQuotesRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedQuotesRule.MatchList(sentences)
}
