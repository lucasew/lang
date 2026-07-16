package nl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DutchUnpairedBracketsRule ports Dutch GenericUnpairedBracketsRule symbol set.
type DutchUnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewDutchUnpairedBracketsRule(messages map[string]string) *DutchUnpairedBracketsRule {
	start := []string{"[", "(", "{", "“", "‹", "“", "„", "\""}
	end := []string{"]", ")", "}", "”", "›", "”", "”", "\""}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	return &DutchUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *DutchUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
