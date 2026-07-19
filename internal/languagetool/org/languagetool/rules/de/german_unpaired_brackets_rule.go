package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanUnpairedBracketsRule ports org.languagetool.rules.de.GermanUnpairedBracketsRule
// (brackets only; quotes are handled by GermanUnpairedQuotesRule).
type GermanUnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewGermanUnpairedBracketsRule(messages map[string]string) *GermanUnpairedBracketsRule {
	start := []string{"[", "(", "{"}
	end := []string{"]", ")", "}"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	// Java: getId returns UNPAIRED_BRACKETS (no DE_ prefix for compatibility)
	if base != nil {
		base.SetRuleID("UNPAIRED_BRACKETS")
		// Java example: unclosed ( before DBV
		base.URL = "https://languagetool.org/insights/de/beitrag/klammern/"
		base.AddExamplePair(
			rules.Wrong("Dem Präsidenten des Deutschen Bauernverbands <marker>(</marker>DBV zufolge habe die Dürre einen Schaden von 1,4 Millionen verursacht."),
			rules.Fixed("Dem Präsidenten des Deutschen Bauernverbands <marker>(</marker>DBV) zufolge habe die Dürre einen Schaden von 1,4 Millionen verursacht."),
		)
	}
	return &GermanUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *GermanUnpairedBracketsRule) GetID() string {
	if r != nil && r.GenericUnpairedBracketsRule != nil {
		return r.GenericUnpairedBracketsRule.GetID()
	}
	return "UNPAIRED_BRACKETS"
}

// GetURL ports GermanUnpairedBracketsRule constructor setUrl.
func (r *GermanUnpairedBracketsRule) GetURL() string {
	return "https://languagetool.org/insights/de/beitrag/klammern/"
}

func (r *GermanUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	// Java attaches this (DE rule) so setUrl is visible on matches.
	ms := r.GenericUnpairedBracketsRule.MatchList(sentences)
	for _, m := range ms {
		if m != nil {
			m.Rule = r
		}
	}
	return ms
}
