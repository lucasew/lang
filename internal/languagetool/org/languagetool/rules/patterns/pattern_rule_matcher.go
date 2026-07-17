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
func (m *PatternRuleMatcher) matchFrom(sentence *languagetool.AnalyzedSentence, tokens []*languagetool.AnalyzedTokenReadings, start int) (*rules.RuleMatch, bool) {
	pos := start
	firstMatch, lastMatch := -1, -1
	firstMarker, lastMarker := -1, -1
	prevSkip := 0

	for ki, matcher := range m.matchers {
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
		// search window: current pos .. pos+prevSkip
		// SkipNext -1 means unlimited (LT PatternToken.skip = -1).
		windowEnd := pos
		if prevSkip < 0 {
			windowEnd = len(tokens) - 1
		} else {
			windowEnd = pos + prevSkip
			if windowEnd >= len(tokens) {
				windowEnd = len(tokens) - 1
			}
		}
		matchedCount := 0
		foundAt := -1
		for try := pos; try <= windowEnd && try < len(tokens) && matchedCount < maxOcc; try++ {
			if tokens[try].IsImmunized() {
				continue
			}
			if matcher.IsMatchedReadings(tokens[try]) {
				if firstMatch < 0 {
					firstMatch = try
				}
				lastMatch = try
				if pt.InsideMarker {
					if firstMarker < 0 {
						firstMarker = try
					}
					lastMarker = try
				}
				// greedy consume consecutive maxOcc from try
				foundAt = try
				matchedCount = 1
				j := try + 1
				for matchedCount < maxOcc && j < len(tokens) {
					if tokens[j].IsImmunized() || !matcher.IsMatchedReadings(tokens[j]) {
						break
					}
					lastMatch = j
					if pt.InsideMarker {
						lastMarker = j
					}
					matchedCount++
					j++
				}
				pos = try + matchedCount
				break
			}
			// for min=0, not finding in window is OK
		}
		if matchedCount < minOcc {
			// optional element: advance without consuming
			if minOcc == 0 {
				prevSkip = pt.SkipNext
				_ = ki
				continue
			}
			return nil, false
		}
		if foundAt < 0 && minOcc == 0 {
			prevSkip = pt.SkipNext
			continue
		}
		prevSkip = pt.SkipNext
	}
	if firstMatch < 0 || lastMatch < 0 {
		return nil, false
	}
	if firstMarker < 0 {
		firstMarker, lastMarker = firstMatch, lastMatch
	}
	fromPos := tokens[firstMarker].GetStartPos()
	toPos := tokens[lastMarker].GetEndPos()
	msg := m.Rule.Message
	if msg == "" {
		msg = m.Rule.Description
	}
	rm := rules.NewRuleMatch(m.Rule, sentence, fromPos, toPos, msg)
	rm.ShortMessage = m.Rule.ShortMessage
	return rm, true
}
