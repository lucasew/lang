package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// UppercaseSentenceStartRule wraps the core UppercaseSentenceStartRule.
// Java German.getRelevantRules passes DE setUrl for uppercase sentence starts.
type UppercaseSentenceStartRule struct {
	*rules.UppercaseSentenceStartRule
}

func NewUppercaseSentenceStartRule(messages map[string]string) *UppercaseSentenceStartRule {
	return &UppercaseSentenceStartRule{UppercaseSentenceStartRule: rules.NewUppercaseSentenceStartRule(messages, "de")}
}

// GetURL ports German.java getRelevantRules URL for UppercaseSentenceStartRule.
func (r *UppercaseSentenceStartRule) GetURL() string {
	return "https://languagetool.org/insights/de/beitrag/gross-klein-schreibung-rechtschreibung/#1-satzanf%C3%A4nge-schreiben-wir-gro%C3%9F"
}

func (r *UppercaseSentenceStartRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	// Java attaches this (DE rule) so setUrl is visible on matches.
	ms := r.UppercaseSentenceStartRule.MatchList(sentences)
	for _, m := range ms {
		if m != nil {
			m.Rule = r
		}
	}
	return ms
}
