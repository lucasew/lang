package sl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// UnpairedBracketsRule wraps GenericUnpairedBracketsRule for this language.
type UnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewUnpairedBracketsRule(messages map[string]string) *UnpairedBracketsRule {
	start := []string{"[", "(", "{", "“", "«", "\"", "'", "‘"}
	end := []string{"]", ")", "}", "”", "»", "\"", "'", "’"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	base.SetRuleID("SL_UNPAIRED_BRACKETS")
	return &UnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *UnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
