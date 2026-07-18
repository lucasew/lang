package patterns

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// MistakeMarker ports PatternRuleMatcher.MISTAKE.
const MistakeMarker = "<mistake/>"

// PatternRuleMatcher ports org.languagetool.rules.patterns.PatternRuleMatcher
// as a sequential token matcher (skip/unification/exceptions deferred).
type PatternRuleMatcher struct {
	Rule     *AbstractTokenBasedRule
	matchers []*PatternTokenMatcher
	// InterpretPreDisambig when true uses pre-disambiguation tokens (not yet wired).
	InterpretPreDisambig bool
}

func NewPatternRuleMatcher(rule *AbstractTokenBasedRule) *PatternRuleMatcher {
	if rule == nil {
		panic("rule required")
	}
	ms := make([]*PatternTokenMatcher, 0, len(rule.Tokens))
	for _, pt := range rule.Tokens {
		ms = append(ms, NewPatternTokenMatcher(pt))
	}
	return &PatternRuleMatcher{Rule: rule, matchers: ms}
}

// NewPatternRuleMatcherFromPattern builds a matcher from a PatternRule.
func NewPatternRuleMatcherFromPattern(rule *PatternRule) *PatternRuleMatcher {
	if rule == nil {
		panic("rule required")
	}
	atr := &AbstractTokenBasedRule{PatternRule: rule}
	return NewPatternRuleMatcher(atr)
}

// Match ports PatternRuleMatcher.match.
func (m *PatternRuleMatcher) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if sentence == nil || m == nil || m.Rule == nil {
		return nil, nil
	}
	// fast path: token hints
	for _, h := range m.Rule.TokenHints {
		if h.CanBeIgnoredFor(sentence) {
			return nil, nil
		}
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	if m.Rule.MinTokenCount > 0 && len(tokens) < m.Rule.MinTokenCount {
		return nil, nil
	}
	var found []*rules.RuleMatch
	patternSize := len(m.matchers)
	if patternSize == 0 {
		return nil, nil
	}
	// limit like Java: tokens.length - patternSize + 1 (+ min occur correction deferred)
	minToks := minPatternTokens(m.matchers)
	limit := len(tokens) - minToks + 1
	if limit < 0 {
		limit = 0
	}
	for i := 0; i < limit; i++ {
		if rm, ok := m.matchFrom(sentence, tokens, i); ok {
			found = append(found, rm)
		}
	}
	return rules.NewRuleWithMaxFilter().Filter(found), nil
}

func minPatternTokens(matchers []*PatternTokenMatcher) int {
	n := 0
	for _, m := range matchers {
		if m.Base == nil {
			n++
			continue
		}
		if m.Base.MinOccurrence > 0 {
			n += m.Base.MinOccurrence
		}
	}
	if n == 0 {
		return 1
	}
	return n
}

// matchFrom tries to match the pattern starting at token index start.
// Optional elements (min=0) backtrack so soft POS over-acceptance does not
// greedily steal tokens needed by later pattern elements (e.g. NL FULL_SENTENCE_001).
func (m *PatternRuleMatcher) matchFrom(sentence *languagetool.AnalyzedSentence, tokens []*languagetool.AnalyzedTokenReadings, start int) (*rules.RuleMatch, bool) {
	type span struct{ first, last, firstMark, lastMark int }
	var rec func(ki, pos, prevSkip int, sp span) (*rules.RuleMatch, bool)
	rec = func(ki, pos, prevSkip int, sp span) (*rules.RuleMatch, bool) {
		if ki >= len(m.matchers) {
			if sp.first < 0 || sp.last < 0 {
				return nil, false
			}
			fm, lm := sp.firstMark, sp.lastMark
			if fm < 0 {
				fm, lm = sp.first, sp.last
			}
			fromPos := tokens[fm].GetStartPos()
			toPos := tokens[lm].GetEndPos()
			msg := m.Rule.Message
			if msg == "" {
				msg = m.Rule.Description
			}
			rm := rules.NewRuleMatch(m.Rule, sentence, fromPos, toPos, msg)
			rm.ShortMessage = m.Rule.ShortMessage
			return rm, true
		}
		matcher := m.matchers[ki]
		pt := matcher.Base
		if pt == nil {
			return nil, false
		}
		minOcc := pt.MinOccurrence
		maxOcc := pt.MaxOccurrence
		if maxOcc < 1 {
			maxOcc = 1
		}
		if minOcc < 0 {
			minOcc = 0
		}
		windowEnd := pos
		if prevSkip < 0 {
			windowEnd = len(tokens) - 1
		} else {
			windowEnd = pos + prevSkip
			if windowEnd >= len(tokens) {
				windowEnd = len(tokens) - 1
			}
		}
		// Try occurrence counts from max down to min (include 0 = skip optional).
		for occ := maxOcc; occ >= minOcc; occ-- {
			if occ == 0 {
				if rm, ok := rec(ki+1, pos, pt.SkipNext, sp); ok {
					return rm, true
				}
				continue
			}
			// Find a start in [pos, windowEnd] that yields occ consecutive matches.
			for try := pos; try <= windowEnd && try < len(tokens); try++ {
				if tokens[try].IsImmunized() || !matcher.IsMatchedReadings(tokens[try]) {
					continue
				}
				// consume occ consecutive
				ok := true
				end := try
				for c := 1; c < occ; c++ {
					j := try + c
					if j >= len(tokens) || tokens[j].IsImmunized() || !matcher.IsMatchedReadings(tokens[j]) {
						ok = false
						break
					}
					end = j
				}
				if !ok {
					continue
				}
				nsp := sp
				if nsp.first < 0 {
					nsp.first = try
				}
				nsp.last = end
				if pt.InsideMarker {
					if nsp.firstMark < 0 {
						nsp.firstMark = try
					}
					nsp.lastMark = end
				}
				if rm, ok := rec(ki+1, end+1, pt.SkipNext, nsp); ok {
					return rm, true
				}
				// try next window position
			}
		}
		return nil, false
	}
	return rec(0, start, 0, span{-1, -1, -1, -1})
}
