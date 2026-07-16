package ga

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// LongParagraphRule wraps the core LongParagraphRule for this language.
type LongParagraphRule struct {
	*rules.LongParagraphRule
}

func NewLongParagraphRule(messages map[string]string, maxWords int) *LongParagraphRule {
	if maxWords <= 0 {
		maxWords = 150
	}
	return &LongParagraphRule{LongParagraphRule: rules.NewLongParagraphRule(messages, maxWords)}
}

func (r *LongParagraphRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.LongParagraphRule.MatchList(sentences)
}
