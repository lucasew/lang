package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// DisambiguationPatternRuleReplacer ports
// org.languagetool.tagging.disambiguation.rules.DisambiguationPatternRuleReplacer
// as a multi-rule applicator over DisambiguationPatternRule.Replace.
type DisambiguationPatternRuleReplacer struct {
	Rules []*DisambiguationPatternRule
}

func NewDisambiguationPatternRuleReplacer(rules []*DisambiguationPatternRule) *DisambiguationPatternRuleReplacer {
	return &DisambiguationPatternRuleReplacer{Rules: append([]*DisambiguationPatternRule(nil), rules...)}
}

// Replace applies all rules in order to the sentence.
func (r *DisambiguationPatternRuleReplacer) Replace(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if r == nil || sentence == nil {
		return sentence
	}
	out := sentence
	for _, rule := range r.Rules {
		if rule == nil {
			continue
		}
		out = rule.Replace(out)
	}
	return out
}

// ReplaceOne applies a single rule (Java instance-per-rule style).
func ReplaceOne(rule *DisambiguationPatternRule, sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if rule == nil {
		return sentence
	}
	return rule.Replace(sentence)
}
