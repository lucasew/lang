package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RussianUnpairedBracketsRule ports org.languagetool.rules.ru.RussianUnpairedBracketsRule.
type RussianUnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewRussianUnpairedBracketsRule(messages map[string]string) *RussianUnpairedBracketsRule {
	// Java RU_START/END symbols; id RU_UNPAIRED_BRACKETS.
	start := []string{"(", "{", "„", "\"", "'", "“"}
	end := []string{")", "}", "“", "\"", "'", "”"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	base.SetRuleID("RU_UNPAIRED_BRACKETS")
	base.AddExamplePair(
		rules.Wrong("Самоотверженный поступок Оленина <marker>(</marker>подарок Лукашке коня вызывает лишь удивление и усиливает недоверие к нему станичников."),
		rules.Fixed("Самоотверженный поступок Оленина <marker>(</marker>подарок Лукашке коня) вызывает лишь удивление и усиливает недоверие к нему станичников."),
	)
	return &RussianUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *RussianUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
