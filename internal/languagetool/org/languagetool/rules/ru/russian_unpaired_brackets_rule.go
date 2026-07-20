package ru

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RussianUnpairedBracketsRule ports org.languagetool.rules.ru.RussianUnpairedBracketsRule.
type RussianUnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

// Java NUMERALS_RU (Cyrillic letter suffixes + Latin/Roman).
var ruNumerals = regexp.MustCompile(`(?i)\d{1,2}?[а-я]*|[а-я]|[А-Я]|[а-я][а-я]|[А-Я][А-Я]|(?i)\d{1,2}?[a-z']*|M*(D?C{0,3}|C[DM])(L?X{0,3}|X[LC])(V?I{0,3}|I[VX])$`)

func NewRussianUnpairedBracketsRule(messages map[string]string) *RussianUnpairedBracketsRule {
	// Java RU_START/END symbols + NUMERALS_RU; id RU_UNPAIRED_BRACKETS.
	start := []string{"(", "{", "„", "\"", "'", "“"}
	end := []string{")", "}", "“", "\"", "'", "”"}
	base := rules.NewGenericUnpairedBracketsRuleWithNumerals(messages, start, end, ruNumerals)
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
