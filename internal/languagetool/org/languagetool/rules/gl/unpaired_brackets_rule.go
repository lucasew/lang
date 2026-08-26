package gl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// UnpairedBracketsRule ports Galician GenericUnpairedBracketsRule symbol set
// (id remains UNPAIRED_BRACKETS — Java does not invent GL_UNPAIRED_BRACKETS).
type UnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewUnpairedBracketsRule(messages map[string]string) *UnpairedBracketsRule {
	// Java Galician.getRelevantRules:
	// "[", "(", "{", "“", "«", "»", "‘", "\"", "'" /
	// "]", ")", "}", "”", "»", "«", "’", "\"", "'"
	start := []string{"[", "(", "{", "“", "«", "»", "‘", "\"", "'"}
	end := []string{"]", ")", "}", "”", "»", "«", "’", "\"", "'"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	return &UnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *UnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
