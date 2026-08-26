package sv

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// UnpairedBracketsRule ports Swedish default GenericUnpairedBracketsRule
// (id UNPAIRED_BRACKETS — Java does not invent SV_UNPAIRED_BRACKETS).
type UnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewUnpairedBracketsRule(messages map[string]string) *UnpairedBracketsRule {
	// Java Swedish: new GenericUnpairedBracketsRule(messages) → default symbols.
	start := []string{"[", "(", "{", "\"", "'"}
	end := []string{"]", ")", "}", "\"", "'"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	return &UnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *UnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
