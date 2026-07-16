package sv

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// WordRepeatRule wraps the core WordRepeatRule for this language.
type WordRepeatRule struct {
	*rules.WordRepeatRule
}

func NewWordRepeatRule(messages map[string]string) *WordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	base.IDOverride = "SV_WORD_REPEAT_RULE"
	return &WordRepeatRule{WordRepeatRule: base}
}

func (r *WordRepeatRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.WordRepeatRule.Match(sentence)
}
