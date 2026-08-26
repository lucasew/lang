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

// NumeralsRUJava is the exact Java NUMERALS_RU Pattern.compile argument
// (RussianUnpairedBracketsRule.NUMERALS_RU).
// Java Matcher.matches() requires a full-string match; Go uses NumeralsRU (anchored).
const NumeralsRUJava = `(?i)\d{1,2}?[а-я]*|[а-я]|[А-Я]|[а-я][а-я]|[А-Я][А-Я]|(?i)\d{1,2}?[a-z']*|M*(D?C{0,3}|C[DM])(L?X{0,3}|X[LC])(V?I{0,3}|I[VX])$`

// NumeralsRU ports Java NUMERALS_RU with ^…$ so regexp.MatchString matches Java Matcher.matches().
// Same approach as GenericUnpairedBracketsRule defaultNumerals (NUMERALS_EN).
var NumeralsRU = regexp.MustCompile(`(?i)^(\d{1,2}?[а-я]*|[а-я]|[А-Я]|[а-я][а-я]|[А-Я][А-Я]|\d{1,2}?[a-z']*|M*(D?C{0,3}|C[DM])(L?X{0,3}|X[LC])(V?I{0,3}|I[VX]))$`)

// NewRussianUnpairedBracketsRule ports
// RussianUnpairedBracketsRule(ResourceBundle messages, Language language).
// Java's language argument is unused (not passed to super).
func NewRussianUnpairedBracketsRule(messages map[string]string) *RussianUnpairedBracketsRule {
	// Java RU_START_SYMBOLS / RU_END_SYMBOLS + NUMERALS_RU
	start := []string{"(", "{", "„", "\"", "'", "“"}
	end := []string{")", "}", "“", "\"", "'", "”"}
	base := rules.NewGenericUnpairedBracketsRuleWithNumerals(messages, start, end, NumeralsRU)
	// Java getId → RU_UNPAIRED_BRACKETS
	base.SetRuleID("RU_UNPAIRED_BRACKETS")
	// Java addExamplePair
	base.AddExamplePair(
		rules.Wrong("Самоотверженный поступок Оленина <marker>(</marker>подарок Лукашке коня вызывает лишь удивление и усиливает недоверие к нему станичников."),
		rules.Fixed("Самоотверженный поступок Оленина <marker>(</marker>подарок Лукашке коня) вызывает лишь удивление и усиливает недоверие к нему станичников."),
	)
	return &RussianUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

// MatchList ports TextLevelRule.match inherited from GenericUnpairedBracketsRule.
// Rebinds Rule on each match to this RU type so getId matches Java `this`.
func (r *RussianUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.GenericUnpairedBracketsRule == nil {
		return nil
	}
	ms := r.GenericUnpairedBracketsRule.MatchList(sentences)
	for _, m := range ms {
		if m != nil {
			m.Rule = r
		}
	}
	return ms
}
