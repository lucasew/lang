package pt

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PortugueseUnpairedBracketsRule wraps GenericUnpairedBracketsRule for PT.
type PortugueseUnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewPortugueseUnpairedBracketsRule(messages map[string]string) *PortugueseUnpairedBracketsRule {
	start := []string{"[", "(", "{", "“", "«", "\"", "'", "‘"}
	end := []string{"]", ")", "}", "”", "»", "\"", "'", "’"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	base.SetRuleID("PT_UNPAIRED_BRACKETS")
	return &PortugueseUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *PortugueseUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
