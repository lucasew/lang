package el

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// UnpairedBracketsRule ports Greek GenericUnpairedBracketsRule("EL_UNPAIRED_BRACKETS", ...).
type UnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewUnpairedBracketsRule(messages map[string]string) *UnpairedBracketsRule {
	// Java Greek.getRelevantRules: "[", "(", "{", "“", "\"", "«" / "]", ")", "}", "”", "\"", "»"
	start := []string{"[", "(", "{", "“", "\"", "«"}
	end := []string{"]", ")", "}", "”", "\"", "»"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	base.SetRuleID("EL_UNPAIRED_BRACKETS")
	return &UnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *UnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
