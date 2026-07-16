package patterns

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// RuleIDGetter is the minimal surface for rules in a RuleSet.
type RuleIDGetter interface {
	GetID() string
}

// RuleSet ports org.languagetool.rules.patterns.RuleSet (plain + id cache).
// Hinted filtering (token/lemma) is deferred until AbstractTokenBasedRule cues exist.
type RuleSet interface {
	AllRules() []RuleIDGetter
	RulesForSentence(sentence *languagetool.AnalyzedSentence) []RuleIDGetter
	AllRuleIDs() map[string]struct{}
}

type plainRuleSet struct {
	rules []RuleIDGetter
	ids   map[string]struct{}
}

// PlainRuleSet returns a set that always yields all rules for any sentence.
func PlainRuleSet(rules []RuleIDGetter) RuleSet {
	ids := map[string]struct{}{}
	for _, r := range rules {
		if r != nil {
			ids[r.GetID()] = struct{}{}
		}
	}
	return &plainRuleSet{rules: append([]RuleIDGetter(nil), rules...), ids: ids}
}

func (p *plainRuleSet) AllRules() []RuleIDGetter { return p.rules }
func (p *plainRuleSet) RulesForSentence(_ *languagetool.AnalyzedSentence) []RuleIDGetter {
	return p.rules
}
func (p *plainRuleSet) AllRuleIDs() map[string]struct{} { return p.ids }
