package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// CatalanUnpairedBracketsRule ports org.languagetool.rules.ca.CatalanUnpairedBracketsRule.
type CatalanUnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewCatalanUnpairedBracketsRule(messages map[string]string) *CatalanUnpairedBracketsRule {
	start := []string{"[", "(", "{", "“", "«", "\"", "'", "‘"}
	end := []string{"]", ")", "}", "”", "»", "\"", "'", "’"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	return &CatalanUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *CatalanUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
