package es

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SpanishUnpairedBracketsRule ports org.languagetool.rules.es.SpanishUnpairedBracketsRule
// using GenericUnpairedBracketsRule (without ES-specific exceptions for D'Hondt etc.).
type SpanishUnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewSpanishUnpairedBracketsRule(messages map[string]string) *SpanishUnpairedBracketsRule {
	start := []string{"[", "(", "{", "“", "«", "\"", "'", "‘"}
	end := []string{"]", ")", "}", "”", "»", "\"", "'", "’"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	return &SpanishUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *SpanishUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
