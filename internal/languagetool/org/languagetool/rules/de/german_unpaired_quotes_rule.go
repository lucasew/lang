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

func (r *GermanUnpairedQuotesRule) GetID() string {
	if r != nil && r.GenericUnpairedQuotesRule != nil {
		return r.GenericUnpairedQuotesRule.GetID()
	}
	return "DE_UNPAIRED_QUOTES"
}

// GetURL ports GermanUnpairedQuotesRule constructor setUrl.
func (r *GermanUnpairedQuotesRule) GetURL() string {
	return "https://languagetool.org/insights/de/beitrag/klammern/"
}

func (r *GermanUnpairedQuotesRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	// Java attaches this (DE rule) so setUrl is visible on matches.
	ms := r.GenericUnpairedQuotesRule.MatchList(sentences)
	for _, m := range ms {
		if m != nil {
			m.Rule = r
		}
	}
	return ms
}
