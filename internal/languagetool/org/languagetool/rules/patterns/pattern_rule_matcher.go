package patterns

import (
	"strings"
	"unicode"
	"unicode/utf8"

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
				// Japanese grammar.xml often writes char/short surfaces (な+い, さ+せる)
				// while Sen/kagome emits one morpheme (ない, させる). Soft bridge:
				// (1) one analysis morph covers several pattern surfaces (concat),
				// (2) one pattern surface covers several analysis morphs (しい←し+い).
				if n := softCJKConcatCover(m.matchers[ki:], tokens[try]); n > 0 {
					nsp := sp
					if nsp.first < 0 {
						nsp.first = try
					}
					nsp.last = try
					for j := 0; j < n; j++ {
						if m.matchers[ki+j].Base != nil && m.matchers[ki+j].Base.InsideMarker {
							if nsp.firstMark < 0 {
								nsp.firstMark = try
							}
							nsp.lastMark = try
						}
					}
					lastSkip := 0
					if m.matchers[ki+n-1].Base != nil {
						lastSkip = m.matchers[ki+n-1].Base.SkipNext
					}
					if rm, ok := rec(ki+n, try+1, lastSkip, nsp); ok {
						return rm, true
					}
				}
				if n := softCJKAnalysisSpan(matcher, tokens[try:]); n > 1 {
					end := try + n - 1
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
				}
			}
		}
		return nil, false
	}
	return rec(0, start, 0, span{-1, -1, -1, -1})
}

// softCJKConcatCover reports how many consecutive pattern tokens are covered by
// one pure-CJK analysis token when their surfaces concatenate to that token.
// Returns 0 if not applicable (non-CJK, POS-only tokens, regexp bodies, etc.).
func softCJKConcatCover(matchers []*PatternTokenMatcher, atr *languagetool.AnalyzedTokenReadings) int {
	if atr == nil || len(matchers) < 2 {
		return 0
	}
	surf := atr.GetToken()
	rs := []rune(surf)
	if len(rs) < 2 || len(rs) > 12 {
		return 0
	}
	for _, r := range rs {
		if !isSoftCJKRune(r) {
			return 0
		}
	}
	built := ""
	n := 0
	for i := 0; i < len(matchers) && n < 8; i++ {
		pt := matchers[i].Base
		if pt == nil || pt.Token == "" {
			return 0
		}
		// Only plain surfaces (or simple | alts without metacharacters).
		part := pt.Token
		if pt.Regexp {
			if strings.ContainsAny(part, ".*+?[](){}\\") {
				return 0
			}
			// Multi-alt RE: try each bar-separated alt for prefix growth.
			alts := strings.Split(part, "|")
			matched := false
			for _, alt := range alts {
				alt = strings.TrimSpace(alt)
				if alt == "" {
					continue
				}
				cand := built + alt
				if cand == surf || strings.HasPrefix(surf, cand) {
					built = cand
					matched = true
					break
				}
			}
			if !matched {
				return 0
			}
		} else {
			// Non-RE: optional case-insensitive for Latin; CJK exact.
			cand := built + part
			if cand != surf && !strings.HasPrefix(surf, cand) {
				// try lowercase latin in part only
				return 0
			}
			built = cand
		}
		n++
		if built == surf {
			// Require at least 2 pattern tokens and prefer multi-token cover
			// over false single-token (already handled by IsMatchedReadings).
			if n >= 2 && utf8.RuneCountInString(built) == len(rs) {
				return n
			}
			return 0
		}
	}
	return 0
}

// softCJKAnalysisSpan reports how many consecutive pure-CJK analysis tokens
// concatenate to the pattern token surface (しい ← し+い). Opposite of
// softCJKConcatCover. Returns 0 if not applicable.
func softCJKAnalysisSpan(matcher *PatternTokenMatcher, tokens []*languagetool.AnalyzedTokenReadings) int {
	if matcher == nil || matcher.Base == nil || len(tokens) < 2 {
		return 0
	}
	pt := matcher.Base
	if pt.Token == "" || pt.Pos != nil && pt.Token == "" {
		return 0
	}
	want := pt.Token
	if pt.Regexp {
		if strings.ContainsAny(want, ".*+?[](){}\\") {
			return 0
		}
		// single alt or bar alts handled below
	}
	alts := []string{want}
	if pt.Regexp && strings.Contains(want, "|") {
		alts = nil
		for _, a := range strings.Split(want, "|") {
			a = strings.TrimSpace(a)
			if a != "" {
				alts = append(alts, a)
			}
		}
	}
	for _, target := range alts {
		trs := []rune(target)
		if len(trs) < 2 {
			continue
		}
		allCJK := true
		for _, r := range trs {
			if !unicode.Is(unicode.Han, r) && !unicode.In(r, unicode.Hiragana, unicode.Katakana) {
				allCJK = false
				break
			}
		}
		if !allCJK {
			continue
		}
		built := ""
		n := 0
		for i := 0; i < len(tokens) && n < 8; i++ {
			if tokens[i] == nil || tokens[i].IsImmunized() {
				break
			}
			t := tokens[i].GetToken()
			for _, r := range t {
				if !isSoftCJKRune(r) {
					return 0
				}
			}
			built += t
			n++
			if built == target {
				if n >= 2 {
					return n
				}
				return 0
			}
			if !strings.HasPrefix(target, built) {
				break
			}
		}
	}
	return 0
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
