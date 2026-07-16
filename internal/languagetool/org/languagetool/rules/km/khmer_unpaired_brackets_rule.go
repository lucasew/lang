package km

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// KhmerUnpairedBracketsRule ports org.languagetool.rules.km.KhmerUnpairedBracketsRule.
type KhmerUnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewKhmerUnpairedBracketsRule(messages map[string]string) *KhmerUnpairedBracketsRule {
	start := []string{"[", "(", "{", "“", "\"", "'", "«"}
	end := []string{"]", ")", "}", "”", "\"", "'", "»"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	base.SetRuleID("KM_UNPAIRED_BRACKETS")
	return &KhmerUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *KhmerUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
