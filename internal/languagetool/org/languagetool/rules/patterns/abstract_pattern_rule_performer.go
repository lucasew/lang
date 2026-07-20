package patterns

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// MatchConsumer is called for each successful pattern match span.
// Indices are into the non-whitespace token array.
type MatchConsumer func(tokenPositions []int, firstMatchToken, lastMatchToken, firstMarker, lastMarker int)

// AbstractPatternRulePerformer ports
// org.languagetool.rules.patterns.AbstractPatternRulePerformer (simplified).
type AbstractPatternRulePerformer struct {
	Rule     *AbstractTokenBasedRule
	Unifier  *Unifier
	matchers []*PatternTokenMatcher
}

func NewAbstractPatternRulePerformer(rule *AbstractTokenBasedRule, unifier *Unifier) *AbstractPatternRulePerformer {
	if rule == nil {
		panic("rule required")
	}
	if unifier == nil {
		unifier = NewUnifier(nil, nil)
	}
	ms := make([]*PatternTokenMatcher, 0, len(rule.Tokens))
	for _, pt := range rule.Tokens {
		ms = append(ms, NewPatternTokenMatcher(pt))
	}
	return &AbstractPatternRulePerformer{Rule: rule, Unifier: unifier, matchers: ms}
}

// DoMatch scans the sentence and invokes consumer for each match (no RuleMatch creation).
// Java AbstractPatternRulePerformer.doMatch: token hints, anchor starts, matchFrom.
// SkipImmunized is false (disambig performer path; PatternRuleMatcher sets true).
func (p *AbstractPatternRulePerformer) DoMatch(sentence *languagetool.AnalyzedSentence, consumer MatchConsumer) {
	if sentence == nil || consumer == nil || len(p.matchers) == 0 {
		return
	}
	if p.Rule != nil {
		for _, h := range p.Rule.TokenHints {
			if h.CanBeIgnoredFor(sentence) {
				return
			}
		}
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	// reuse PatternRuleMatcher for limit/anchor/matchFromResult (Java inheritance)
	prm := &PatternRuleMatcher{Rule: p.Rule, matchers: p.matchers, SkipImmunized: false}
	limit := prm.matchStartLimit(len(tokens))
	starts := prm.matchStartIndices(sentence, limit)
	for _, i := range starts {
		res, ok := prm.matchFromResult(sentence, tokens, i)
		if !ok || res == nil {
			continue
		}
		consumer(res.Positions, res.First, res.Last, res.FirstMark, res.LastMark)
	}
}
