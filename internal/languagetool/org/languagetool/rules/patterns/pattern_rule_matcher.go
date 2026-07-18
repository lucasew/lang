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

// NewPatternRuleMatcherStrict builds a matcher with StrictPOS on every token
// (Java disambiguation: unknown surfaces only satisfy postag=UNKNOWN).
func NewPatternRuleMatcherStrict(rule *AbstractTokenBasedRule) *PatternRuleMatcher {
	m := NewPatternRuleMatcher(rule)
	for _, mt := range m.matchers {
		if mt != nil {
			mt.StrictPOS = true
		}
	}
	return m
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
		// Java PatternToken max="-1" means unlimited occurrences (LOOK_DOOR
		// min=0 max=-1 chunk_re="[BI]-NP.*"). Soft previously clamped max<1 to 1.
		if maxOcc < 0 {
			// cap by remaining tokens so the occ loop stays finite
			maxOcc = len(tokens) - pos
			if maxOcc < minOcc {
				maxOcc = minOcc
			}
		} else if maxOcc < 1 {
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
					// Java: scope="previous" exception blocks when previous token matches.
					if pt.HasPreviousException() && try > 0 &&
						matcher.IsMatchedByPreviousException(tokens[try-1]) {
						continue
					}
					// Java: scope="next" exception blocks when following token matches
					// (also when skip=0 and next is the immediate neighbor).
					if pt.HasNextException() && try+1 < len(tokens) &&
						matcher.IsMatchedByNextException(tokens[try+1]) {
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
				// Soft DOUBLE_BANG etc.: UK/JA keep "!!" as one token while soft
				// grammar patterns use <token>!</token><token>!</token>.
				if n := softRepeatedPunctCover(m.matchers[ki:], tokens[try]); n > 0 {
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
				// Soft: pattern "week-end" / "e-mail" when tokenizer splits week + - + end.
				if n := softHyphenatedSurfaceCover(matcher, tokens[try:]); n > 1 {
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
				// Soft: pattern "al"/"del" when CatalanWordTokenizer splits a+l / de+l.
				if n := softFusedPrepCover(matcher, tokens[try:]); n > 1 {
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

// softHyphenatedSurfaceCover: pattern token "week-end" matches analysis week + - + end
// (FrenchWordTokenizer splits hyphens; soft FR-CA regional rules use full compounds).
func softHyphenatedSurfaceCover(matcher *PatternTokenMatcher, tokens []*languagetool.AnalyzedTokenReadings) int {
	if matcher == nil || matcher.Base == nil || len(tokens) < 3 {
		return 0
	}
	pt := matcher.Base
	if pt.Token == "" || pt.Regexp || !strings.Contains(pt.Token, "-") {
		return 0
	}
	want := strings.ToLower(pt.Token)
	parts := strings.Split(want, "-")
	if len(parts) < 2 {
		return 0
	}
	// Expect part, "-", part, "-", ... alternating
	ti := 0
	for pi, part := range parts {
		if ti >= len(tokens) || tokens[ti] == nil || tokens[ti].IsImmunized() {
			return 0
		}
		if !strings.EqualFold(tokens[ti].GetToken(), part) {
			return 0
		}
		ti++
		if pi < len(parts)-1 {
			if ti >= len(tokens) || tokens[ti] == nil || tokens[ti].GetToken() != "-" {
				return 0
			}
			ti++
		}
	}
	if ti < 3 {
		return 0
	}
	return ti
}

// softFusedPrepCover: pattern "al"/"del"/"pel" matches tokenizer splits a+l, de+l, pe+l
// (CatalanWordTokenizer pe(ls?) pattern; soft picky rules use fused orthography).
func softFusedPrepCover(matcher *PatternTokenMatcher, tokens []*languagetool.AnalyzedTokenReadings) int {
	if matcher == nil || matcher.Base == nil || len(tokens) < 2 {
		return 0
	}
	pt := matcher.Base
	if pt.Token == "" || pt.Regexp {
		return 0
	}
	want := strings.ToLower(pt.Token)
	var parts []string
	switch want {
	case "al":
		parts = []string{"a", "l"}
	case "als":
		parts = []string{"a", "ls"}
	case "del":
		parts = []string{"de", "l"}
	case "dels":
		parts = []string{"de", "ls"}
	case "pel":
		parts = []string{"pe", "l"}
	case "pels":
		parts = []string{"pe", "ls"}
	case "al'", "a-l":
		return 0
	default:
		return 0
	}
	for i, p := range parts {
		if i >= len(tokens) || tokens[i] == nil || tokens[i].IsImmunized() {
			return 0
		}
		if !strings.EqualFold(tokens[i].GetToken(), p) {
			return 0
		}
	}
	return len(parts)
}

// softRepeatedPunctCover reports how many consecutive single-char pattern tokens
// are covered by one analysis token of the same repeated punctuation (e.g. "!!"
// covers !+!). Used by soft DOUBLE_BANG packs; Ukrainian/Japanese tokenizers keep
// !{2,3} as one token (Java UkrainianWordTokenizer SPLIT_CHARS multi-punct).
func softRepeatedPunctCover(matchers []*PatternTokenMatcher, atr *languagetool.AnalyzedTokenReadings) int {
	if atr == nil || len(matchers) < 2 {
		return 0
	}
	surf := atr.GetToken()
	rs := []rune(surf)
	if len(rs) < 2 || len(rs) > 4 {
		return 0
	}
	base := rs[0]
	if unicode.IsLetter(base) || unicode.IsDigit(base) || unicode.IsSpace(base) {
		return 0
	}
	for _, r := range rs[1:] {
		if r != base {
			return 0
		}
	}
	want := string(base)
	n := 0
	for i := 0; i < len(matchers) && n < len(rs); i++ {
		pt := matchers[i].Base
		if pt == nil || pt.Token != want || pt.Regexp {
			break
		}
		n++
	}
	if n == len(rs) && n >= 2 {
		return n
	}
	return 0
}
