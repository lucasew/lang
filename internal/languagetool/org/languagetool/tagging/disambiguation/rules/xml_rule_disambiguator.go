package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// XmlRuleDisambiguator ports
// org.languagetool.tagging.disambiguation.rules.XmlRuleDisambiguator
// with an injected rule list (XML loading via DisambiguationRuleLoader later).
type XmlRuleDisambiguator struct {
	disambiguation.AbstractDisambiguator
	Rules         []*DisambiguationPatternRule
	UnifierConfig *patterns.UnifierConfiguration
	// ruleSet ports Java RuleSet.textHinted(disambiguationRulesList) for
	// rulesForSentence filtering (token-hint skip without changing semantics).
	ruleSet patterns.RuleSet
}

func NewXmlRuleDisambiguator(rules []*DisambiguationPatternRule) *XmlRuleDisambiguator {
	list := append([]*DisambiguationPatternRule(nil), rules...)
	// Java: disambiguationRules = RuleSet.textHinted(disambiguationRulesList)
	asID := make([]patterns.RuleIDGetter, 0, len(list))
	for _, r := range list {
		if r != nil {
			asID = append(asID, r)
		}
	}
	return &XmlRuleDisambiguator{
		Rules:   list,
		ruleSet: patterns.TextHintedRuleSet(asID),
	}
}

// Disambiguate applies each disambiguation pattern rule in order.
func (d *XmlRuleDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	return d.DisambiguateWithCancel(input, nil)
}

// DisambiguateWithCancel ports disambiguate with cancellation callback.
// Java: for (Rule rule : disambiguationRules.rulesForSentence(sentence)) replace.
func (d *XmlRuleDisambiguator) DisambiguateWithCancel(sentence *languagetool.AnalyzedSentence, checkCanceled languagetool.CheckCancelledCallback) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	var candidates []patterns.RuleIDGetter
	if d.ruleSet != nil {
		candidates = d.ruleSet.RulesForSentence(sentence)
	} else {
		for _, rule := range d.Rules {
			if rule != nil {
				candidates = append(candidates, rule)
			}
		}
	}
	for _, rg := range candidates {
		if checkCanceled != nil && checkCanceled() {
			break
		}
		rule, ok := rg.(*DisambiguationPatternRule)
		if !ok || rule == nil {
			continue
		}
		if rule.UnifierConfig == nil && d.UnifierConfig != nil {
			rule.UnifierConfig = d.UnifierConfig
		}
		sentence = rule.Replace(sentence)
	}
	return sentence
}

var _ disambiguation.Disambiguator = (*XmlRuleDisambiguator)(nil)
