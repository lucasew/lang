package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanUnpairedBracketsRule ports org.languagetool.rules.de.GermanUnpairedBracketsRule
// (brackets only; quotes are handled by GermanUnpairedQuotesRule).
type GermanUnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewGermanUnpairedBracketsRule(messages map[string]string) *GermanUnpairedBracketsRule {
	start := []string{"[", "(", "{"}
	end := []string{"]", ")", "}"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	return &GermanUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *GermanUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
