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

// HintableRule is a rule that can be skipped via token/lemma cues.
type HintableRule interface {
	RuleIDGetter
	CanBeIgnoredFor(sentence *languagetool.AnalyzedSentence) bool
}

// TextHintedRuleSet excludes rules whose non-inflected token hints miss the sentence.
// Ports RuleSet.textHinted.
func TextHintedRuleSet(rules []RuleIDGetter) RuleSet {
	return newHintedRuleSet(rules, false)
}

// TextLemmaHintedRuleSet excludes rules whose token or lemma hints miss the sentence.
// Ports RuleSet.textLemmaHinted.
func TextLemmaHintedRuleSet(rules []RuleIDGetter) RuleSet {
	return newHintedRuleSet(rules, true)
}

type hintedRuleSet struct {
	rules          []RuleIDGetter
	ids            map[string]struct{}
	withLemmaHints bool
}

func newHintedRuleSet(rules []RuleIDGetter, withLemmaHints bool) *hintedRuleSet {
	ids := map[string]struct{}{}
	out := make([]RuleIDGetter, 0, len(rules))
	for _, r := range rules {
		if r == nil {
			continue
		}
		out = append(out, r)
		ids[r.GetID()] = struct{}{}
	}
	return &hintedRuleSet{rules: out, ids: ids, withLemmaHints: withLemmaHints}
}

func (h *hintedRuleSet) AllRules() []RuleIDGetter               { return h.rules }
func (h *hintedRuleSet) AllRuleIDs() map[string]struct{}        { return h.ids }
func (h *hintedRuleSet) RulesForSentence(s *languagetool.AnalyzedSentence) []RuleIDGetter {
	if s == nil {
		return h.rules
	}
	var out []RuleIDGetter
	for _, r := range h.rules {
		if hr, ok := r.(HintableRule); ok {
			// withLemmaHints: use full CanBeIgnoredFor (token+lemma hints)
			// text-only: if rule only has lemma hints, treat as unclassified (always include)
			if !h.withLemmaHints {
				if atr, ok := r.(*AbstractTokenBasedRule); ok && onlyInflectedHints(atr) {
					out = append(out, r)
					continue
				}
			}
			if hr.CanBeIgnoredFor(s) {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

func onlyInflectedHints(r *AbstractTokenBasedRule) bool {
	if r == nil || len(r.TokenHints) == 0 {
		return false
	}
	for _, th := range r.TokenHints {
		if !th.Inflected {
			return false
		}
	}
	return true
}
