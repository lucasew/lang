package sk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// WordRepeatBeginningRule wraps the core WordRepeatBeginningRule for this language.
type WordRepeatBeginningRule struct {
	*rules.WordRepeatBeginningRule
}

func NewWordRepeatBeginningRule(messages map[string]string) *WordRepeatBeginningRule {
	base := rules.NewWordRepeatBeginningRule(messages)
	base.IDOverride = "SK_WORD_REPEAT_BEGINNING_RULE"
	return &WordRepeatBeginningRule{WordRepeatBeginningRule: base}
}

func (r *WordRepeatBeginningRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.WordRepeatBeginningRule.MatchList(sentences)
}
