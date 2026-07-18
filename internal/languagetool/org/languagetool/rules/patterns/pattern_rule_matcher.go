package patterns

import (
	"strings"
	"unicode"

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
	// Java: tokens.length - patternSize + 1. Soft CJK cover maps several pattern
	// tokens onto fewer analysis morphs (and often only SENT_START + one word),
	// so always try every start index; matchFrom fails quickly on mismatches.
	limit := len(tokens)
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
				// Optional element absent: preserve the previous token's skip window
				// so later required elements still see skip="N" (Java PatternRuleMatcher).
				// Using pt.SkipNext here would drop e.g. couper skip=4 before dépenses.
				if rm, ok := rec(ki+1, pos, prevSkip, sp); ok {
					return rm, true
				}
				continue
			}
			// Find a start in [pos, windowEnd] that yields occ consecutive matches.
			for try := pos; try <= windowEnd && try < len(tokens); try++ {
				if tokens[try].IsImmunized() {
					continue
				}
				if matcher.IsMatchedReadings(tokens[try]) {
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
					continue
				}
				if occ != 1 {
					continue
				}
				// Japanese grammar.xml often uses short surfaces that cut across
				// Sen/kagome morph boundaries (最高+調 vs 最+高調, ず+らい vs ずら+い).
				// Soft: match a run of pattern surfaces as consecutive substrings of
				// the pure-CJK analysis stream from try (Java remains 1:1 morph match).
				if nPat, nAna := softCJKSurfaceAlign(m.matchers[ki:], tokens[try:]); nPat > 0 {
					end := try + nAna - 1
					nsp := sp
					if nsp.first < 0 {
						nsp.first = try
					}
					nsp.last = end
					for j := 0; j < nPat; j++ {
						if m.matchers[ki+j].Base != nil && m.matchers[ki+j].Base.InsideMarker {
							if nsp.firstMark < 0 {
								nsp.firstMark = try
							}
							nsp.lastMark = end
						}
					}
					lastSkip := 0
					if m.matchers[ki+nPat-1].Base != nil {
						lastSkip = m.matchers[ki+nPat-1].Base.SkipNext
					}
					if rm, ok := rec(ki+nPat, try+nAna, lastSkip, nsp); ok {
						return rm, true
					}
				}
			}
		}
		return nil, false
	}
	return rec(0, start, 0, span{-1, -1, -1, -1})
}

// softCJKSurfaceAlign matches a prefix of pattern tokens as consecutive substrings
// of the pure-CJK analysis stream starting at tokens[0].
// Returns (nPattern, nAnalysis). Both 0 if not applicable.
//
// Examples (kagome analysis → grammar surfaces):
//
//	最+高調  → 最高+調
//	ずら+い  → ず+らい
//	おそ+る  → お+そる
//	させる   → さ+せる
func softCJKSurfaceAlign(matchers []*PatternTokenMatcher, tokens []*languagetool.AnalyzedTokenReadings) (nPat, nAna int) {
	if len(matchers) == 0 || len(tokens) == 0 {
		return 0, 0
	}
	// Flatten a pure-CJK prefix of analysis tokens into a rune stream with
	// reverse map from rune index → analysis token index.
	var stream []rune
	var runeToTok []int
	for i, atr := range tokens {
		if atr == nil || atr.IsImmunized() {
			break
		}
		s := atr.GetToken()
		if s == "" {
			break
		}
		ok := true
		for _, r := range s {
			if !isSoftCJKRune(r) {
				ok = false
				break
			}
		}
		if !ok {
			break
		}
		for _, r := range s {
			stream = append(stream, r)
			runeToTok = append(runeToTok, i)
		}
	}
	if len(stream) == 0 {
		return 0, 0
	}
	pos := 0
	for mi := 0; mi < len(matchers) && mi < 12; mi++ {
		pt := matchers[mi].Base
		if pt == nil || pt.Token == "" {
			break
		}
		// POS-only or complex RE: stop soft surface align
		if pt.Regexp && strings.ContainsAny(pt.Token, ".*+?[](){}\\") {
			break
		}
		alts := []string{pt.Token}
		if pt.Regexp && strings.Contains(pt.Token, "|") {
			alts = nil
			for _, a := range strings.Split(pt.Token, "|") {
				a = strings.TrimSpace(a)
				if a != "" {
					alts = append(alts, a)
				}
			}
		}
		matched := false
		for _, alt := range alts {
			ar := []rune(alt)
			if len(ar) == 0 || pos+len(ar) > len(stream) {
				continue
			}
			if string(stream[pos:pos+len(ar)]) == alt {
				pos += len(ar)
				nPat++
				matched = true
				break
			}
		}
		if !matched {
			break
		}
	}
	if nPat == 0 || pos == 0 {
		return 0, 0
	}
	nAna = runeToTok[pos-1] + 1
	// Skip pure 1:1 single-token (already handled by IsMatchedReadings).
	if nPat == 1 && nAna == 1 {
		return 0, 0
	}
	return nPat, nAna
}

// isSoftCJKRune is true for Han/kana and the prolonged sound mark ー (U+30FC),
// which Go's unicode.Katakana table does not include but Sen/kagome keeps inside
// katakana words (ストリート, …).
func isSoftCJKRune(r rune) bool {
	if unicode.Is(unicode.Han, r) || unicode.In(r, unicode.Hiragana, unicode.Katakana) {
		return true
	}
	// ー prolonged sound; iteration marks sometimes appear in soft JA packs
	return r == 'ー' || r == 'ゝ' || r == 'ゞ' || r == 'ヽ' || r == 'ヾ'
}
