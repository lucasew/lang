package crh

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// EmptyLineRule wraps the core EmptyLineRule for this language.
type EmptyLineRule struct {
	*rules.EmptyLineRule
}

func NewEmptyLineRule(messages map[string]string) *EmptyLineRule {
	return &EmptyLineRule{EmptyLineRule: rules.NewEmptyLineRule(messages)}
}

func (r *EmptyLineRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.EmptyLineRule.MatchList(sentences)
}
