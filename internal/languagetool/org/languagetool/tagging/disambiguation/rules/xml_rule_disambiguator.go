package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// XmlRuleDisambiguator ports
// org.languagetool.tagging.disambiguation.rules.XmlRuleDisambiguator
// with an injected rule list (XML loading via DisambiguationRuleLoader later).
type XmlRuleDisambiguator struct {
	disambiguation.AbstractDisambiguator
	Rules []*DisambiguationPatternRule
}

func NewXmlRuleDisambiguator(rules []*DisambiguationPatternRule) *XmlRuleDisambiguator {
	return &XmlRuleDisambiguator{Rules: append([]*DisambiguationPatternRule(nil), rules...)}
}

// Disambiguate applies each disambiguation pattern rule in order.
func (d *XmlRuleDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	return d.DisambiguateWithCancel(input, nil)
}

// DisambiguateWithCancel ports disambiguate with cancellation callback.
func (d *XmlRuleDisambiguator) DisambiguateWithCancel(sentence *languagetool.AnalyzedSentence, checkCanceled languagetool.CheckCancelledCallback) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	for _, rule := range d.Rules {
		if checkCanceled != nil && checkCanceled() {
			break
		}
		if rule == nil {
			continue
		}
		sentence = rule.Replace(sentence)
	}
	return sentence
}

var _ disambiguation.Disambiguator = (*XmlRuleDisambiguator)(nil)
