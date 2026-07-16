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
func (p *AbstractPatternRulePerformer) DoMatch(sentence *languagetool.AnalyzedSentence, consumer MatchConsumer) {
	if sentence == nil || consumer == nil || len(p.matchers) == 0 {
		return
	}
	for _, h := range p.Rule.TokenHints {
		if h.CanBeIgnoredFor(sentence) {
			return
		}
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	minToks := 0
	for _, m := range p.matchers {
		if m.Base != nil && m.Base.MinOccurrence > 0 {
			minToks += m.Base.MinOccurrence
		}
	}
	if minToks == 0 {
		minToks = 1
	}
	limit := len(tokens) - minToks + 1
	if limit < 0 {
		return
	}
	// reuse PatternRuleMatcher matchFrom via a temporary matcher
	prm := &PatternRuleMatcher{Rule: p.Rule, matchers: p.matchers}
	for i := 0; i < limit; i++ {
		rm, ok := prm.matchFrom(sentence, tokens, i)
		if !ok || rm == nil {
			continue
		}
		// recover indices from positions
		first, last := -1, -1
		for ti, t := range tokens {
			if t.GetStartPos() == rm.FromPos {
				first = ti
			}
			if t.GetEndPos() == rm.ToPos {
				last = ti
			}
		}
		if first < 0 || last < 0 {
			// approximate by range
			for ti, t := range tokens {
				if t.GetStartPos() >= rm.FromPos && first < 0 {
					first = ti
				}
				if t.GetStartPos() < rm.ToPos {
					last = ti
				}
			}
		}
		if first < 0 || last < 0 {
			continue
		}
		positions := make([]int, last-first+1)
		for j := range positions {
			positions[j] = 1
		}
		consumer(positions, first, last, first, last)
	}
}
