package ro

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RomanianUnpairedBracketsRule ports Romanian GenericUnpairedBracketsRule symbols.
type RomanianUnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewRomanianUnpairedBracketsRule(messages map[string]string) *RomanianUnpairedBracketsRule {
	start := []string{"[", "(", "{", "„", "«", "»"}
	end := []string{"]", ")", "}", "”", "»", "«"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	return &RomanianUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *RomanianUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
