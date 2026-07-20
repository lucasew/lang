package patterns

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// MistakeMarker ports PatternRuleMatcher.MISTAKE.
const MistakeMarker = "<mistake/>"

// suggestionStartTag / suggestionEndTag port RuleMatch.SUGGESTION_*_TAG.
const (
	suggestionStartTag = "<suggestion>"
	suggestionEndTag   = "</suggestion>"
)

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
	// Java AbstractPatternRulePerformer.doMatch start enumeration.
	limit := m.matchStartLimit(len(tokens))
	starts := m.matchStartIndices(sentence, limit)
	for _, i := range starts {
		if rm, ok := m.matchFrom(sentence, tokens, i); ok {
			found = append(found, rm)
		}
	}
	return rules.NewRuleWithMaxFilter().Filter(found), nil
}

// matchStartLimit ports doMatch limit:
// isSentStart ? 1 : max(0, tokens.length - patternSize + 1) + minOccurCorrection
func (m *PatternRuleMatcher) matchStartLimit(tokenCount int) int {
	if m == nil {
		return 0
	}
	if m.Rule != nil && m.Rule.IsSentStart() {
		return 1
	}
	patternSize := len(m.matchers)
	minOccurCorrection := 0
	for _, mt := range m.matchers {
		if mt != nil && mt.Base != nil && mt.Base.MinOccurrence == 0 {
			minOccurCorrection++
		}
	}
	limit := tokenCount - patternSize + 1
	if limit < 0 {
		limit = 0
	}
	return limit + minOccurCorrection
}

// matchStartIndices ports doMatch anchor vs full scan.
// When anchorHint is set and not raw_pos, only try anchor-derived starts
// (Java: if anchorIndices != null). Empty after filter means no starts.
func (m *PatternRuleMatcher) matchStartIndices(sentence *languagetool.AnalyzedSentence, limit int) []int {
	if limit <= 0 {
		return nil
	}
	if m.Rule != nil && m.Rule.AnchorHint != nil && !m.usePreDisambigTokens() {
		idxs := m.Rule.AnchorHint.GetPossibleIndices(sentence)
		if idxs != nil {
			var starts []int
			for _, ai := range idxs {
				i := ai - m.Rule.AnchorHint.TokenIndex
				if i >= 0 && i < limit {
					starts = append(starts, i)
				}
			}
			return starts
		}
	}
	starts := make([]int, limit)
	for i := 0; i < limit; i++ {
		starts[i] = i
	}
	return starts
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

// matchResult ports one successful doMatch consumer invocation.
type matchResult struct {
	RM                               *rules.RuleMatch
	Positions                        []int
	First, Last, FirstMark, LastMark int
}

// matchFrom tries to match the pattern starting at token index start.
// Optional elements (min=0) backtrack so soft POS over-acceptance does not
// greedily steal tokens needed by later pattern elements (e.g. NL FULL_SENTENCE_001).
func (m *PatternRuleMatcher) matchFrom(sentence *languagetool.AnalyzedSentence, tokens []*languagetool.AnalyzedTokenReadings, start int) (*rules.RuleMatch, bool) {
	res, ok := m.matchFromResult(sentence, tokens, start)
	if !ok || res == nil {
		return nil, false
	}
	return res.RM, true
}

func (m *PatternRuleMatcher) matchFromResult(sentence *languagetool.AnalyzedSentence, tokens []*languagetool.AnalyzedTokenReadings, start int) (*matchResult, bool) {
	type span struct {
		first, last, firstMark, lastMark int
		// positions[i] = tokens consumed by pattern element i (Java tokenPositions).
		positions []int
	}
	needUni := m.needsUnification()
	var rec func(ki, pos, prevSkip int, sp span, bag *unifyBag) (*matchResult, bool)
	rec = func(ki, pos, prevSkip int, sp span, bag *unifyBag) (*matchResult, bool) {
		if ki >= len(m.matchers) {
			if sp.first < 0 || sp.last < 0 {
				return nil, false
			}
			if needUni && !m.testUnification(bag) {
				return nil, false
			}
			positions := sp.positions
			if len(positions) == 0 {
				positions = defaultPositions(len(m.matchers))
			}
			rm := m.createRuleMatch(sentence, tokens, positions, sp.first, sp.last, sp.firstMark, sp.lastMark)
			if rm == nil {
				return nil, false
			}
			fm, lm := sp.firstMark, sp.lastMark
			if fm < 0 {
				fm, lm = sp.first, sp.last
			}
			return &matchResult{
				RM: rm, Positions: positions,
				First: sp.first, Last: sp.last, FirstMark: fm, LastMark: lm,
			}, true
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
				if res, ok := rec(ki+1, pos, prevSkip, nsp, bag); ok {
					return res, true
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
				// prevSkipNext > 0: prevElement.isMatchedByScopeNextException(each reading of current)
				// → reject token position if any reading hits prev's scope=next exception.
				if prevSkip > 0 && ki > 0 {
					if prevM := m.matchers[ki-1]; prevM != nil && prevM.Base != nil &&
						prevM.Base.HasNextException() &&
						prevM.IsMatchedByNextException(tokens[try]) {
						continue
					}
				}
				// prevSkipNext == 0: current matcher's scope=next vs next token's first reading only
				// (Java: tokens[tokenNo+1].getAnalyzedToken(0)), before accepting the match.
				if prevSkip == 0 && pt.HasNextException() && try+1 < len(tokens) &&
					matcher.IsMatchedByNextExceptionFirstReading(tokens[try+1]) {
					continue
				}
				// When first element, Java still has firstMatchToken=-1 until match;
				// re-resolve with try as provisional first if needed for refs on later elems only.
				if matcher.IsMatchedReadings(tokens[try]) {
					// Java: scope="previous" exception blocks when previous token matches
					// (after anyMatched, still reject).
					if pt.HasPreviousException() && try > 0 &&
						matcher.IsMatchedByPreviousException(tokens[try-1]) {
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
					// Java: tokenPositions[i] = skipShift + 1 where
					// skipShift = lastMatchToken - nextPos (nextPos is search start `pos`).
					// Includes tokens skipped in the skip window before this match.
					consumed := end - pos + 1
					if consumed < 1 {
						consumed = 1
					}
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
					if res, ok := rec(ki+1, end+1, pt.SkipNext, nsp, nbag); ok {
						return res, true
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

// createRuleMatch ports PatternRuleMatcher.createRuleMatch.
func (m *PatternRuleMatcher) createRuleMatch(
	sentence *languagetool.AnalyzedSentence,
	tokens []*languagetool.AnalyzedTokenReadings,
	tokenPositions []int,
	firstMatchToken, lastMatchToken, firstMarkerMatchToken, lastMarkerMatchToken int,
) *rules.RuleMatch {
	if m == nil || m.Rule == nil || firstMatchToken < 0 || lastMatchToken < 0 {
		return nil
	}
	pr := m.Rule.PatternRule
	lang := ""
	var sugMatches, sugMatchesOut []*Match
	msg := m.Rule.Message
	shortMsg := m.Rule.ShortMessage
	sugOut := ""
	startCorr := 0
	adjustCase := true
	if pr != nil {
		lang = pr.LanguageCode
		sugMatches = pr.SuggestionMatches
		sugMatchesOut = pr.SuggestionMatchesOutMsg
		sugOut = pr.SuggestionsOutMsg
		startCorr = pr.StartPositionCorrection
		if pr.AdjustSuggestionCase != nil {
			adjustCase = *pr.AdjustSuggestionCase
		}
		if msg == "" {
			msg = pr.Description
		}
	}
	if msg == "" {
		msg = m.Rule.Description
	}

	errMessage := FormatMatches(tokens, tokenPositions, firstMatchToken, msg, sugMatches, lang)
	shortErrMessage := ""
	if shortMsg != "" {
		shortErrMessage = FormatMatches(tokens, tokenPositions, firstMatchToken, shortMsg, sugMatches, lang)
	}
	suggestionsOutMsg := FormatMatches(tokens, tokenPositions, firstMatchToken, sugOut, sugMatchesOut, lang)

	// startPositionCorrection shifts the case-sample token (Java correctedStPos).
	correctedStPos := 0
	if startCorr > 0 {
		lim := startCorr
		if lim > len(tokenPositions)-1 {
			lim = len(tokenPositions) - 1
		}
		for l := 0; l <= lim && l < len(tokenPositions); l++ {
			correctedStPos += tokenPositions[l]
		}
		correctedStPos--
	}
	idx := firstMatchToken + correctedStPos
	if idx >= len(tokens) {
		idx = len(tokens) - 1
	}
	if idx < 0 {
		idx = 0
	}
	firstMatchTokenObj := tokens[idx]

	// All-uppercase / starts-with-uppercase for suggestion casing.
	var inputTokens []string
	for i := idx; i <= lastMatchToken && i < len(tokens); i++ {
		if tokens[i] != nil {
			inputTokens = append(inputTokens, tokens[i].GetToken())
		}
	}
	isInputAllUppercase := isAllUppercaseTokenList(inputTokens)
	isAllUppercase := isInputAllUppercase &&
		(len(strings.ReplaceAll(firstMatchTokenObj.GetToken(), "'", "")) > 1 || lastMatchToken > idx) &&
		matchPreservesCase(sugMatches, msg) &&
		matchPreservesCase(sugMatchesOut, sugOut)
	isAllUppercase = isAllUppercase && adjustCase

	startsWithUppercase := tools.StartsWithUppercase(firstMatchTokenObj.GetToken()) &&
		matchPreservesCase(sugMatches, msg) &&
		matchPreservesCase(sugMatchesOut, sugOut)
	if firstMatchTokenObj.IsSentenceStart() && firstMatchToken+correctedStPos+1 < len(tokens) {
		firstMatchTokenObj = tokens[firstMatchToken+correctedStPos+1]
		startsWithUppercase = tools.StartsWithUppercase(firstMatchTokenObj.GetToken())
	}
	startsWithUppercase = startsWithUppercase && adjustCase

	if firstMarkerMatchToken < 0 {
		firstMarkerMatchToken = firstMatchToken
	}
	if lastMarkerMatchToken < 0 {
		lastMarkerMatchToken = lastMatchToken
	}
	if firstMarkerMatchToken >= len(tokens) {
		firstMarkerMatchToken = len(tokens) - 1
	}
	if lastMarkerMatchToken >= len(tokens) {
		lastMarkerMatchToken = len(tokens) - 1
	}

	fromPos := tokens[firstMarkerMatchToken].GetStartPos()
	// Java: comma suggestion may extend fromPos back over previous token end.
	if firstMarkerMatchToken >= 1 &&
		(strings.Contains(errMessage, suggestionStartTag+",") ||
			strings.Contains(suggestionsOutMsg, suggestionStartTag+",")) {
		fromPos = tokens[firstMarkerMatchToken-1].GetEndPos()
	}
	toPos := tokens[lastMarkerMatchToken].GetEndPos()
	if fromPos >= toPos {
		return nil
	}

	// suppress_misspelled with no suggestions → no RuleMatch
	if strings.Contains(errMessage, PleaseSpellMe) &&
		!strings.Contains(errMessage, suggestionStartTag) &&
		!strings.Contains(suggestionsOutMsg, suggestionStartTag) {
		return nil
	}
	clearMsg := strings.ReplaceAll(errMessage, PleaseSpellMe, "")
	clearMsg = strings.ReplaceAll(clearMsg, MistakeMarker, "")

	patFrom := tokens[firstMatchToken].GetStartPos()
	patTo := tokens[lastMatchToken].GetEndPos()
	rm := rules.NewRuleMatch(m.Rule, sentence, fromPos, toPos, clearMsg)
	rm.ShortMessage = shortErrMessage
	rm.SetPatternPosition(patFrom, patTo)
	if orig := sentenceTextSlice(sentence, fromPos, toPos); orig != "" {
		rm.OriginalErrorStr = orig
	}

	// Extract <suggestion> from message + suggestionsOutMsg (Java RuleMatch ctor).
	combined := clearMsg + suggestionsOutMsg
	var replacements []string
	for _, rep := range extractSuggestionBodies(combined) {
		if rep == "" || strings.Contains(rep, MistakeMarker) {
			continue
		}
		// Strip residual pleasespellme inside body
		rep = strings.ReplaceAll(rep, PleaseSpellMe, "")
		// Java RuleMatch ctor case adjustment
		if isAllUppercase && !(tools.IsMixedCase(rep) && !strings.Contains(rep, " ")) {
			if rm.OriginalErrorStr != strings.ToUpper(rep) {
				rep = strings.ToUpper(rep)
			}
		} else if startsWithUppercase {
			rep = tools.UppercaseFirstChar(rep)
		}
		replacements = append(replacements, rep)
	}

	// Templates not already covered by message markup (loader path).
	if pr != nil && len(pr.SuggestionTemplates) > 0 && len(replacements) == 0 {
		for _, t := range pr.SuggestionTemplates {
			for _, e := range ExpandSuggestionTemplate(t, tokens, tokenPositions, firstMatchToken, sugMatches, lang) {
				if e == "" || e == MistakeMarker || strings.Contains(e, MistakeMarker) {
					continue
				}
				if suppressMisspelledIn(sugMatches) && isParenOnlyForm(e) {
					continue
				}
				if isAllUppercase {
					up := strings.ToUpper(e)
					if rm.OriginalErrorStr == up {
						continue
					}
					e = up
				} else if startsWithUppercase {
					e = tools.UppercaseFirstChar(e)
				}
				replacements = append(replacements, e)
			}
		}
	}
	if len(replacements) > 0 {
		rm.SetSuggestedReplacements(replacements)
	}

	if pr != nil && pr.Filter != nil {
		patternTokens := tokens[firstMatchToken : lastMatchToken+1]
		eval := NewRuleFilterEvaluator(pr.Filter)
		rm = eval.RunFilter(pr.FilterArgs, rm, patternTokens, firstMatchToken, tokenPositions)
	}
	return rm
}

// matchPreservesCase ports PatternRuleMatcher.matchPreservesCase.
func matchPreservesCase(suggestionMatches []*Match, msg string) bool {
	if len(suggestionMatches) == 0 || msg == "" {
		return true
	}
	sugStart := strings.Index(msg, suggestionStartTag)
	if sugStart < 0 {
		return true
	}
	sugStart += len(suggestionStartTag)
	if strings.Contains(msg, PleaseSpellMe) {
		// Java adds PLEASE_SPELL_ME length when present in message
		// only if it appears at suggestion start region — approximate:
		if idx := strings.Index(msg[sugStart:], PleaseSpellMe); idx == 0 {
			sugStart += len(PleaseSpellMe)
		}
	}
	if sugStart >= len(msg) {
		return true
	}
	for _, sMatch := range suggestionMatches {
		if sMatch == nil || sMatch.IsInMessageOnly() {
			continue
		}
		if sMatch.ConvertsCase() && msg[sugStart] == '\\' {
			return false
		}
	}
	return true
}

func isAllUppercaseTokenList(tokens []string) bool {
	if len(tokens) == 0 {
		return false
	}
	isInputAllUppercase := true
	isAllNotLetters := true
	for _, s := range tokens {
		isInputAllUppercase = isInputAllUppercase && tools.IsAllUppercase(s)
		isAllNotLetters = isAllNotLetters && (tools.IsNotWordString(s) || tools.IsPunctuationMark(s))
	}
	return isInputAllUppercase && !isAllNotLetters
}

func extractSuggestionBodies(s string) []string {
	var out []string
	for {
		start := strings.Index(s, suggestionStartTag)
		if start < 0 {
			break
		}
		rest := s[start+len(suggestionStartTag):]
		end := strings.Index(rest, suggestionEndTag)
		if end < 0 {
			break
		}
		out = append(out, rest[:end])
		s = rest[end+len(suggestionEndTag):]
	}
	return out
}

func sentenceTextSlice(sentence *languagetool.AnalyzedSentence, from, to int) string {
	if sentence == nil {
		return ""
	}
	text := sentence.GetText()
	// Positions are UTF-16-oriented in Java; for BMP text byte==UTF-16 for ASCII.
	// Use rune-safe slice by mapping via token positions when possible.
	if from < 0 {
		from = 0
	}
	if to > len(text) {
		to = len(text)
	}
	if from >= to {
		return ""
	}
	// Prefer byte indices when they fall inside the Go string (AnalyzePlain uses UTF-16-friendly for BMP).
	if to <= len(text) {
		return text[from:to]
	}
	return ""
}
