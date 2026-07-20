package patterns

import (
	"strings"

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
	// InterpretPreDisambig when true uses pre-disambiguation tokens (raw_pos).
	InterpretPreDisambig bool
	// SkipImmunized when true, immunized tokens cannot participate in a match
	// (Java PatternRuleMatcher.testAllReadings). Disambiguation uses AbstractPatternRulePerformer
	// which does not skip immunized tokens — set false via NewPatternRuleMatcherStrict.
	SkipImmunized bool
}

func NewPatternRuleMatcher(rule *AbstractTokenBasedRule) *PatternRuleMatcher {
	if rule == nil {
		panic("rule required")
	}
	ms := make([]*PatternTokenMatcher, 0, len(rule.Tokens))
	for _, pt := range rule.Tokens {
		ms = append(ms, NewPatternTokenMatcher(pt))
	}
	// Java PatternRuleMatcher skips immunized tokens.
	return &PatternRuleMatcher{Rule: rule, matchers: ms, SkipImmunized: true}
}

// NewPatternRuleMatcherStrict builds a matcher with StrictPOS on every token
// (Java disambiguation: unknown surfaces only satisfy postag=UNKNOWN).
// SkipImmunized is false — Java AbstractPatternRulePerformer (disambig) does not
// skip immunized tokens, so overlapping IMMUNIZE anti-patterns can complete.
func NewPatternRuleMatcherStrict(rule *AbstractTokenBasedRule) *PatternRuleMatcher {
	m := NewPatternRuleMatcher(rule)
	m.SkipImmunized = false
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
	// Java PatternRule extends AbstractTokenBasedRule: constructor computes tokenHints / minTokenCount.
	atr := &AbstractTokenBasedRule{PatternRule: rule}
	atr.computeHints(rule.Tokens)
	m := NewPatternRuleMatcher(atr)
	m.InterpretPreDisambig = rule.InterpretPreDisambig
	return m
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
	// Java: raw_pos → pre-disambiguation tokens (PatternRuleMatcher.match).
	tokens := sentence.GetTokensWithoutWhitespace()
	if m.usePreDisambigTokens() {
		if pre := sentence.GetPreDisambigTokensWithoutWhitespace(); len(pre) > 0 {
			tokens = pre
		}
	}
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

// unifyBag collects matched readings for AbstractPatternRulePerformer.testUnification.
type unifyBag struct {
	toUnify map[*PatternToken][][]*languagetool.AnalyzedToken
	neutral map[*PatternToken][]*languagetool.AnalyzedTokenReadings
}

func newUnifyBag() *unifyBag {
	return &unifyBag{
		toUnify: map[*PatternToken][][]*languagetool.AnalyzedToken{},
		neutral: map[*PatternToken][]*languagetool.AnalyzedTokenReadings{},
	}
}

func (b *unifyBag) clone() *unifyBag {
	if b == nil {
		return nil
	}
	out := newUnifyBag()
	for k, sets := range b.toUnify {
		cp := make([][]*languagetool.AnalyzedToken, len(sets))
		for i, s := range sets {
			cp[i] = append([]*languagetool.AnalyzedToken(nil), s...)
		}
		out.toUnify[k] = cp
	}
	for k, atrs := range b.neutral {
		out.neutral[k] = append([]*languagetool.AnalyzedTokenReadings(nil), atrs...)
	}
	return out
}

func (b *unifyBag) record(matcher *PatternTokenMatcher, pt *PatternToken, atrs []*languagetool.AnalyzedTokenReadings) {
	if b == nil || pt == nil || matcher == nil {
		return
	}
	if pt.IsUnificationNeutral() {
		for _, atr := range atrs {
			if atr != nil {
				b.neutral[pt] = append(b.neutral[pt], atr)
			}
		}
		return
	}
	if !pt.IsUnified() {
		return
	}
	for _, atr := range atrs {
		if atr == nil {
			continue
		}
		readings := matcher.CollectMatchedReadings(atr)
		if len(readings) > 0 {
			b.toUnify[pt] = append(b.toUnify[pt], readings)
		}
	}
}

// usePreDisambigTokens ports isInterpretPosTagsPreDisambiguation.
func (m *PatternRuleMatcher) usePreDisambigTokens() bool {
	if m == nil {
		return false
	}
	if m.InterpretPreDisambig {
		return true
	}
	return m.Rule != nil && m.Rule.PatternRule != nil && m.Rule.PatternRule.InterpretPreDisambig
}

// needsUnification reports whether any pattern token participates in unify.
func (m *PatternRuleMatcher) needsUnification() bool {
	for _, mt := range m.matchers {
		if mt != nil && mt.Base != nil && mt.Base.IsUnified() {
			return true
		}
	}
	return false
}

// testUnification ports AbstractPatternRulePerformer.testUnification.
// Fail-closed: rules with <unify> and no UnifierConfig never match.
func (m *PatternRuleMatcher) testUnification(bag *unifyBag) bool {
	if !m.needsUnification() {
		return true
	}
	var cfg *UnifierConfiguration
	if m.Rule != nil && m.Rule.PatternRule != nil {
		cfg = m.Rule.PatternRule.UnifierConfig
	}
	if cfg == nil || bag == nil {
		// Without equivalence tables, uniNegated would false-fire — refuse.
		return false
	}
	uni := cfg.CreateUnifier()
	for _, matcher := range m.matchers {
		if matcher == nil || matcher.Base == nil {
			continue
		}
		pt := matcher.Base
		if neutrals, ok := bag.neutral[pt]; ok && len(neutrals) > 0 {
			for _, atr := range neutrals {
				uni.AddNeutralElement(atr)
			}
			continue
		}
		readingSets := bag.toUnify[pt]
		if len(readingSets) == 0 {
			continue
		}
		for si, readings := range readingSets {
			anyMatched := false
			for i, reading := range readings {
				lastReading := i == len(readings)-1
				anyMatched = anyMatched || uni.IsUnified(reading, pt.GetUniFeatures(), lastReading)
			}
			// Empty reading set: still need lastReading semantics for empty?
			// Java only iterates non-empty lists collected from matches.
			if pt.IsUniNegated() && anyMatched {
				return false
			}
			if pt.IsLastInUnification() && si == len(readingSets)-1 {
				if !anyMatched && !pt.IsUniNegated() {
					return false
				}
				uni.Reset()
			}
		}
	}
	return true
}

// matchFrom tries to match the pattern starting at token index start.
// Optional elements (min=0) backtrack so soft POS over-acceptance does not
// greedily steal tokens needed by later pattern elements (e.g. NL FULL_SENTENCE_001).
func (m *PatternRuleMatcher) matchFrom(sentence *languagetool.AnalyzedSentence, tokens []*languagetool.AnalyzedTokenReadings, start int) (*rules.RuleMatch, bool) {
	type span struct {
		first, last, firstMark, lastMark int
		// positions[i] = tokens consumed by pattern element i (Java tokenPositions).
		positions []int
	}
	needUni := m.needsUnification()
	var rec func(ki, pos, prevSkip int, sp span, bag *unifyBag) (*rules.RuleMatch, bool)
	rec = func(ki, pos, prevSkip int, sp span, bag *unifyBag) (*rules.RuleMatch, bool) {
		if ki >= len(m.matchers) {
			if sp.first < 0 || sp.last < 0 {
				return nil, false
			}
			if needUni && !m.testUnification(bag) {
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
			// Java createRuleMatch: formatMatches on message/short before RuleMatch.
			lang := ""
			var sugMatches []*Match
			if m.Rule.PatternRule != nil {
				lang = m.Rule.PatternRule.LanguageCode
				sugMatches = m.Rule.PatternRule.SuggestionMatches
			}
			positions := sp.positions
			if len(positions) == 0 {
				positions = defaultPositions(len(m.matchers))
			}
			msg = FormatMatches(tokens, positions, sp.first, msg, sugMatches, lang)
			shortMsg := m.Rule.ShortMessage
			if shortMsg != "" {
				shortMsg = FormatMatches(tokens, positions, sp.first, shortMsg, sugMatches, lang)
			}
			rm := rules.NewRuleMatch(m.Rule, sentence, fromPos, toPos, msg)
			rm.ShortMessage = shortMsg
			// Expand suggestion templates (Java formatMatches on suggestion markup).
			if m.Rule.PatternRule != nil && len(m.Rule.PatternRule.SuggestionTemplates) > 0 {
				var expanded []string
				for _, t := range m.Rule.PatternRule.SuggestionTemplates {
					for _, e := range ExpandSuggestionTemplate(t, tokens, positions, sp.first, sugMatches, lang) {
						// Java removeSuppressMisspelled: drop misspelled / non-synth markers.
						if e == "" || e == MistakeMarker || strings.Contains(e, MistakeMarker) {
							continue
						}
						// Empty synth form "(word)" under suppress_misspelled is dropped by
						// removeSuppressMisspelled when wrapped in suggestion tags; templates
						// are already unwrapped — drop parenthesized-only forms when any
						// match was suppress_misspelled.
						if suppressMisspelledIn(sugMatches) && isParenOnlyForm(e) {
							continue
						}
						expanded = append(expanded, e)
					}
				}
				if len(expanded) > 0 {
					rm.SetSuggestedReplacements(expanded)
				}
			}
			// Java PatternRuleMatcher.createRuleMatch: run RuleFilter when set.
			if m.Rule.PatternRule != nil && m.Rule.PatternRule.Filter != nil {
				patternTokens := tokens[sp.first : sp.last+1]
				eval := NewRuleFilterEvaluator(m.Rule.PatternRule.Filter)
				rm = eval.RunFilter(m.Rule.PatternRule.FilterArgs, rm, patternTokens, sp.first, positions)
				if rm == nil {
					return nil, false
				}
			}
			return rm, true
		}
		matcher := m.matchers[ki]
		pt := matcher.Base
		if pt == nil {
			return nil, false
		}
		// Java AbstractPatternRulePerformer: resolveReference + prepareAndGroup before testing.
		lang := ""
		if m.Rule != nil && m.Rule.PatternRule != nil {
			lang = m.Rule.PatternRule.LanguageCode
		}
		matcher.ResolveReference(sp.first, tokens, lang)
		matcher.PrepareAndGroup(sp.first, tokens, lang)
		pt = matcher.GetPatternToken()
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
				nsp := sp
				nsp.positions = append(append([]int(nil), sp.positions...), 0)
				if rm, ok := rec(ki+1, pos, prevSkip, nsp, bag); ok {
					return rm, true
				}
				continue
			}
			// Find a start in [pos, windowEnd] that yields occ consecutive matches.
			for try := pos; try <= windowEnd && try < len(tokens); try++ {
				// Java PatternRuleMatcher skips immunized; AbstractPatternRulePerformer (disambig) does not.
				if m.SkipImmunized && tokens[try].IsImmunized() {
					continue
				}
				// Java AbstractPatternRulePerformer.testAllReadings prevMatched:
				// when previous element had skip>0, its scope=next exception rejects
				// candidate tokens in the skip window that match the exception.
				if prevSkip > 0 && ki > 0 {
					if prevM := m.matchers[ki-1]; prevM != nil && prevM.Base != nil &&
						prevM.Base.HasNextException() &&
						prevM.IsMatchedByNextException(tokens[try]) {
						continue
					}
				}
				// When first element, Java still has firstMatchToken=-1 until match;
				// re-resolve with try as provisional first if needed for refs on later elems only.
				if matcher.IsMatchedReadings(tokens[try]) {
					// Java: scope="previous" exception blocks when previous token matches.
					if pt.HasPreviousException() && try > 0 &&
						matcher.IsMatchedByPreviousException(tokens[try-1]) {
						continue
					}
					// Java: scope="next" on *this* element vs immediate next token only when
					// prevSkipNext == 0 (skip>0 uses the prevMatched path above for following elems).
					if prevSkip == 0 && pt.HasNextException() && try+1 < len(tokens) &&
						matcher.IsMatchedByNextException(tokens[try+1]) {
						continue
					}
					// consume occ consecutive
					ok := true
					end := try
					for c := 1; c < occ; c++ {
						j := try + c
						if j >= len(tokens) || (m.SkipImmunized && tokens[j].IsImmunized()) || !matcher.IsMatchedReadings(tokens[j]) {
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
					consumed := end - try + 1
					nsp.positions = append(append([]int(nil), sp.positions...), consumed)
					nbag := bag
					if needUni {
						nbag = bag.clone()
						if nbag == nil {
							nbag = newUnifyBag()
						}
						var spanAtrs []*languagetool.AnalyzedTokenReadings
						for j := try; j <= end; j++ {
							spanAtrs = append(spanAtrs, tokens[j])
						}
						nbag.record(matcher, pt, spanAtrs)
					}
					if rm, ok := rec(ki+1, end+1, pt.SkipNext, nsp, nbag); ok {
						return rm, true
					}
					continue
				}
				if occ != 1 {
					continue
				}
				// Faithful: one pattern token ↔ one analysis token (Java).
				// No soft multi-token cover invents (CJK align, fused prep, etc.).
			}
		}
		return nil, false
	}
	var startBag *unifyBag
	if needUni {
		startBag = newUnifyBag()
	}
	return rec(0, start, 0, span{first: -1, last: -1, firstMark: -1, lastMark: -1}, startBag)
}
