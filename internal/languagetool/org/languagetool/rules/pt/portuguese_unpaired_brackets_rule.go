package pt

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PortugueseUnpairedBracketsRule ports Portuguese GenericUnpairedBracketsRule
// (id remains UNPAIRED_BRACKETS — Java does not invent PT_UNPAIRED_BRACKETS).
type PortugueseUnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewPortugueseUnpairedBracketsRule(messages map[string]string) *PortugueseUnpairedBracketsRule {
	// Java Portuguese.getRelevantRules: "[", "(", "{", "\"", "“" / "]", ")", "}", "\"", "”"
	start := []string{"[", "(", "{", "\"", "“"}
	end := []string{"]", ")", "}", "\"", "”"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	return &PortugueseUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *PortugueseUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
