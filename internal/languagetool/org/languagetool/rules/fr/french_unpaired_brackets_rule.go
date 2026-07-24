package fr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// FrenchUnpairedBracketsRule ports French GenericUnpairedBracketsRule symbols
// (brackets only — French dialog may span sentences).
type FrenchUnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewFrenchUnpairedBracketsRule(messages map[string]string) *FrenchUnpairedBracketsRule {
	start := []string{"[", "(", "{"}
	end := []string{"]", ")", "}"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	return &FrenchUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *FrenchUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
