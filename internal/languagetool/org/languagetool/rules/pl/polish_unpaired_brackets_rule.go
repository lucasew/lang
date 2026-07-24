package pl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PolishUnpairedBracketsRule ports org.languagetool.rules.pl.PolishUnpairedBracketsRule.
type PolishUnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewPolishUnpairedBracketsRule(messages map[string]string) *PolishUnpairedBracketsRule {
	start := []string{"[", "(", "{", "„", "»", "\""}
	end := []string{"]", ")", "}", "”", "«", "\""}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	base.SetRuleID("PL_UNPAIRED_BRACKETS")
	// Java: unpaired „ → close with ”
	base.AddExamplePair(
		rules.Wrong("To jest zdanie z <marker>„</marker>cudzysłowem."),
		rules.Fixed("To jest zdanie z <marker>„</marker>cudzysłowem”."),
	)
	return &PolishUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *PolishUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
