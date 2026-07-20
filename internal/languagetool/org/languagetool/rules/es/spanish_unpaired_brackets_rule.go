package es

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SpanishUnpairedBracketsRule ports org.languagetool.rules.es.SpanishUnpairedBracketsRule.
type SpanishUnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewSpanishUnpairedBracketsRule(messages map[string]string) *SpanishUnpairedBracketsRule {
	// Java ES_START/END symbols.
	start := []string{"[", "(", "{", "“", "«", "\"", "'", "‘"}
	end := []string{"]", ")", "}", "”", "»", "\"", "'", "’"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	base.SetRuleID("ES_UNPAIRED_BRACKETS")
	return &SpanishUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *SpanishUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
