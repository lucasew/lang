package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RussianUnpairedBracketsRule ports org.languagetool.rules.ru.RussianUnpairedBracketsRule
// using GenericUnpairedBracketsRule (numeral exceptions use the default Latin pattern).
type RussianUnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewRussianUnpairedBracketsRule(messages map[string]string) *RussianUnpairedBracketsRule {
	start := []string{"(", "{", "„", "\"", "'", "“"}
	end := []string{")", "}", "“", "\"", "'", "”"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	return &RussianUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *RussianUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
