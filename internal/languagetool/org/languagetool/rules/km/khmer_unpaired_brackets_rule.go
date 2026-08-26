package km

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// KhmerUnpairedBracketsRule ports org.languagetool.rules.km.KhmerUnpairedBracketsRule.
type KhmerUnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

// NewKhmerUnpairedBracketsRule ports
// KhmerUnpairedBracketsRule(ResourceBundle messages, Language language).
// Java's language argument is unused (not passed to super).
func NewKhmerUnpairedBracketsRule(messages map[string]string) *KhmerUnpairedBracketsRule {
	// Java KM_START_SYMBOLS / KM_END_SYMBOLS
	start := []string{"[", "(", "{", "“", "\"", "'", "«"}
	end := []string{"]", ")", "}", "”", "\"", "'", "»"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	// Java getId → KM_UNPAIRED_BRACKETS
	base.SetRuleID("KM_UNPAIRED_BRACKETS")
	return &KhmerUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

// MatchList ports TextLevelRule.match inherited from GenericUnpairedBracketsRule.
// Rebinds Rule on each match to this KM type so getId matches Java `this`.
func (r *KhmerUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.GenericUnpairedBracketsRule == nil {
		return nil
	}
	ms := r.GenericUnpairedBracketsRule.MatchList(sentences)
	for _, m := range ms {
		if m != nil {
			m.Rule = r
		}
	}
	return ms
}
