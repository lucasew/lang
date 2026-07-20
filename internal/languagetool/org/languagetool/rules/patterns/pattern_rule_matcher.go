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
// (extends AbstractPatternRulePerformer.matchFrom sequential algorithm).
type PatternRuleMatcher struct {
	Rule     *AbstractTokenBasedRule
	matchers []*PatternTokenMatcher
	// InterpretPreDisambig when true uses pre-disambiguation tokens (raw_pos).
	InterpretPreDisambig bool
	// SkipImmunized when true, immunized tokens cannot participate in a match
	// (Java PatternRuleMatcher.testAllReadings). Disambiguation uses AbstractPatternRulePerformer
	// which does not skip immunized tokens — set false via NewPatternRuleMatcherStrict.
	SkipImmunized bool
	// UseList ports PatternRuleMatcher.useList — phrase elementNo translation.
	UseList bool
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
	// Java: new PatternRuleMatcher(this, useList)
	if !rule.UseList && len(rule.Tokens) > 0 {
		// Ensure elementNo is computed if rule was mutated after NewPatternRule.
		rule.computeElementNo()
	}
	atr := &AbstractTokenBasedRule{PatternRule: rule}
	atr.computeHints(rule.Tokens)
	m := NewPatternRuleMatcher(atr)
	m.InterpretPreDisambig = rule.InterpretPreDisambig
	m.UseList = rule.UseList
	return m
}

// translateElementNo ports PatternRuleMatcher.translateElementNo.
// When useList, skip values refer to XML-level elements; sum ElementNo for token span.
func (m *PatternRuleMatcher) translateElementNo(i int) int {
	if m == nil || !m.UseList || i < 0 {
		return i
	}
	var elementNo []int
	if m.Rule != nil && m.Rule.PatternRule != nil {
		elementNo = m.Rule.PatternRule.ElementNo
	}
	j := 0
	for k := 0; k < i && k < len(elementNo); k++ {
		j += elementNo[k]
	}
	return j
}

// phraseLen ports PatternRuleMatcher.phraseLen.
func (m *PatternRuleMatcher) phraseLen(i int) int {
	if m == nil || !m.UseList {
		return 1
	}
	var elementNo []int
	if m.Rule != nil && m.Rule.PatternRule != nil {
		elementNo = m.Rule.PatternRule.ElementNo
	}
	if i < 0 || i >= len(elementNo) {
		return 1
	}
	return elementNo[i]
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
// When GetUnified, outUnified receives Unifier.getFinalUnified() before reset.
func (m *PatternRuleMatcher) testUnification(bag *unifyBag) (ok bool, unified []*languagetool.AnalyzedTokenReadings) {
	if !m.needsUnification() {
		return true, nil
	}
	var cfg *UnifierConfiguration
	wantUnified := false
	if m.Rule != nil && m.Rule.PatternRule != nil {
		cfg = m.Rule.PatternRule.UnifierConfig
		wantUnified = m.Rule.PatternRule.GetUnified
	}
	if cfg == nil || bag == nil {
		// Without equivalence tables, uniNegated would false-fire — refuse.
		return false, nil
	}
	uni := cfg.CreateUnifier()
	var lastUnified []*languagetool.AnalyzedTokenReadings
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
				return false, nil
			}
			if pt.IsLastInUnification() && si == len(readingSets)-1 {
				if !anyMatched && !pt.IsUniNegated() {
					return false, nil
				}
				// Java: if (rule.isGetUnified()) unifiedTokens = unifier.getFinalUnified()
				if wantUnified {
					lastUnified = uni.GetFinalUnified()
				}
				uni.Reset()
			}
		}
	}
	return true, lastUnified
}

// matchResult ports one successful doMatch consumer invocation.
type matchResult struct {
	RM                               *rules.RuleMatch
	Positions                        []int
	First, Last, FirstMark, LastMark int
	// UnifiedTokens ports AbstractPatternRulePerformer.unifiedTokens when getUnified.
	UnifiedTokens []*languagetool.AnalyzedTokenReadings
}

// matchFrom tries to match the pattern starting at token index start.
func (m *PatternRuleMatcher) matchFrom(sentence *languagetool.AnalyzedSentence, tokens []*languagetool.AnalyzedTokenReadings, start int) (*rules.RuleMatch, bool) {
	res, ok := m.matchFromResult(sentence, tokens, start)
	if !ok || res == nil || res.RM == nil {
		return nil, false
	}
	return res.RM, true
}

// matchFromResult ports AbstractPatternRulePerformer.matchFrom (sequential greedy).
// Optional min=0 uses Java foundNext look-ahead (not free backtracking).
func (m *PatternRuleMatcher) matchFromResult(sentence *languagetool.AnalyzedSentence, tokens []*languagetool.AnalyzedTokenReadings, start int) (*matchResult, bool) {
	patternSize := len(m.matchers)
	if patternSize == 0 || start < 0 || start >= len(tokens) {
		return nil, false
	}
	lang := ""
	if m.Rule != nil && m.Rule.PatternRule != nil {
		lang = m.Rule.PatternRule.LanguageCode
	}
	tokenPositions := make([]int, patternSize)
	var bag *unifyBag
	if m.needsUnification() {
		bag = newUnifyBag()
	}
	minOccurCorr := m.minOccurCorrection()

	var pTokenMatcher, prevTokenMatcher *PatternTokenMatcher
	skipShiftTotal := 0
	allElementsMatch := false
	matchingTokens := 0
	firstMatchToken := -1
	lastMatchToken := -1
	firstMarkerMatchToken := -1
	lastMarkerMatchToken := -1
	prevSkipNext := 0
	minOccurSkip := 0

	for k := 0; k < patternSize; k++ {
		prevTokenMatcher = pTokenMatcher
		pTokenMatcher = m.matchers[k]
		if pTokenMatcher == nil {
			return nil, false
		}
		pTokenMatcher.ResolveReference(firstMatchToken, tokens, lang)
		nextPos := start + k + skipShiftTotal - minOccurSkip
		// Java: if prevSkipNext + nextPos >= tokens.length || prevSkipNext < 0
		if prevSkipNext+nextPos >= len(tokens) || prevSkipNext < 0 {
			prevSkipNext = len(tokens) - (nextPos + 1)
		}
		maxTok := nextPos + prevSkipNext
		// tokens.length - (patternSize - k) + minOccurCorrection
		capTok := len(tokens) - (patternSize - k) + minOccurCorr
		if maxTok > capTok {
			maxTok = capTok
		}
		if maxTok >= len(tokens) {
			maxTok = len(tokens) - 1
		}
		allElementsMatch = false
		foundOptionalSkip := false
		for mm := nextPos; mm <= maxTok && mm < len(tokens); mm++ {
			allElementsMatch = m.testAllReadings(tokens, pTokenMatcher, prevTokenMatcher, mm, firstMatchToken, prevSkipNext, bag)

			pt := pTokenMatcher.GetPatternToken()
			if pt != nil && pt.MinOccurrence == 0 {
				foundNext := false
				for k2 := k + 1; k2 < patternSize; k2++ {
					nextElement := m.matchers[k2]
					if nextElement == nil {
						continue
					}
					// Java does not resolveReference on look-ahead elements.
					nextMatch := m.testAllReadings(tokens, nextElement, pTokenMatcher, mm, firstMatchToken, prevSkipNext, nil)
					if nextMatch {
						// optional absent: next element wants this token
						allElementsMatch = true
						minOccurSkip++
						if matchingTokens < patternSize {
							tokenPositions[matchingTokens] = 0
							matchingTokens++
						}
						foundNext = true
						foundOptionalSkip = true
						break
					}
					npt := nextElement.GetPatternToken()
					if npt != nil && npt.MinOccurrence > 0 {
						break
					}
				}
				if foundNext {
					break
				}
			}

			if allElementsMatch {
				remainingElems := patternSize - k - 1
				skipForMax := m.skipMaxTokens(tokens, pTokenMatcher, firstMatchToken, prevSkipNext, prevTokenMatcher, mm, remainingElems, bag)
				lastMatchToken = mm + skipForMax
				skipShift := lastMatchToken - nextPos
				if matchingTokens < patternSize {
					tokenPositions[matchingTokens] = skipShift + 1
					matchingTokens++
				}
				if pt == nil {
					pt = pTokenMatcher.GetPatternToken()
				}
				skipNext := 0
				if pt != nil {
					skipNext = pt.SkipNext
				}
				prevSkipNext = m.translateElementNo(skipNext)
				skipShiftTotal += skipShift
				if firstMatchToken == -1 {
					firstMatchToken = lastMatchToken - skipForMax
				}
				if pt != nil && pt.InsideMarker {
					if firstMarkerMatchToken == -1 {
						firstMarkerMatchToken = lastMatchToken - skipForMax
					}
					lastMarkerMatchToken = lastMatchToken
				}
				// Unify already recorded in testAllReadings / skipMaxTokens via bag.
				break
			}
		}
		if foundOptionalSkip {
			// optional skipped with positions=0; continue to next pattern element
			continue
		}
		if !allElementsMatch {
			return nil, false
		}
	}

	if !allElementsMatch || matchingTokens != patternSize {
		return nil, false
	}
	var unified []*languagetool.AnalyzedTokenReadings
	if m.needsUnification() {
		ok, u := m.testUnification(bag)
		if !ok {
			return nil, false
		}
		unified = u
	}
	if firstMatchToken < 0 || lastMatchToken < 0 {
		return nil, false
	}
	positions := append([]int(nil), tokenPositions...)
	// createRuleMatch may return nil (e.g. suppress_misspelled); still a structural match
	// for AbstractPatternRulePerformer.DoMatch (disambig). Match() requires RM != nil.
	rm := m.createRuleMatch(sentence, tokens, positions, firstMatchToken, lastMatchToken, firstMarkerMatchToken, lastMarkerMatchToken)
	fm, lm := firstMarkerMatchToken, lastMarkerMatchToken
	if fm < 0 {
		fm, lm = firstMatchToken, lastMatchToken
	}
	return &matchResult{
		RM: rm, Positions: positions,
		First: firstMatchToken, Last: lastMatchToken, FirstMark: fm, LastMark: lm,
		UnifiedTokens: unified,
	}, true
}

// minOccurCorrection ports AbstractPatternRulePerformer.getMinOccurrenceCorrection.
func (m *PatternRuleMatcher) minOccurCorrection() int {
	n := 0
	for _, mt := range m.matchers {
		if mt != nil && mt.Base != nil && mt.Base.MinOccurrence == 0 {
			n++
		}
	}
	return n
}

// testAllReadings ports AbstractPatternRulePerformer.testAllReadings (+ PatternRuleMatcher immunized).
// bag may be nil (look-ahead checks must not pollute unification).
func (m *PatternRuleMatcher) testAllReadings(
	tokens []*languagetool.AnalyzedTokenReadings,
	matcher, prevElement *PatternTokenMatcher,
	tokenNo, firstMatchToken, prevSkipNext int,
	bag *unifyBag,
) bool {
	if matcher == nil || tokenNo < 0 || tokenNo >= len(tokens) || tokens[tokenNo] == nil {
		return false
	}
	// Java PatternRuleMatcher.testAllReadings: immunized → false
	if m.SkipImmunized && tokens[tokenNo].IsImmunized() {
		return false
	}
	// prevMatched: prev skip window + scope=next on current readings
	if prevSkipNext > 0 && prevElement != nil && prevElement.Base != nil &&
		prevElement.Base.HasNextException() &&
		prevElement.IsMatchedByNextException(tokens[tokenNo]) {
		return false
	}
	// prevSkipNext == 0: current matcher's scope=next vs next token first reading
	pt := matcher.GetPatternToken()
	if prevSkipNext == 0 && pt != nil && pt.HasNextException() && tokenNo+1 < len(tokens) &&
		matcher.IsMatchedByNextExceptionFirstReading(tokens[tokenNo+1]) {
		return false
	}

	lang := ""
	if m.Rule != nil && m.Rule.PatternRule != nil {
		lang = m.Rule.PatternRule.LanguageCode
	}
	matcher.PrepareAndGroup(firstMatchToken, tokens, lang)

	if !matcher.IsMatchedReadings(tokens[tokenNo]) {
		return false
	}
	// scope=previous after anyMatched (Java)
	if pt != nil && pt.HasPreviousException() && tokenNo > 0 &&
		matcher.IsMatchedByPreviousException(tokens[tokenNo-1]) {
		return false
	}
	if bag != nil && pt != nil {
		bag.record(matcher, pt, []*languagetool.AnalyzedTokenReadings{tokens[tokenNo]})
	}
	return true
}

// skipMaxTokens ports AbstractPatternRulePerformer.skipMaxTokens.
func (m *PatternRuleMatcher) skipMaxTokens(
	tokens []*languagetool.AnalyzedTokenReadings,
	elem *PatternTokenMatcher,
	firstMatchToken, prevSkipNext int,
	prevElement *PatternTokenMatcher,
	mm, remainingElems int,
	bag *unifyBag,
) int {
	if elem == nil || elem.GetPatternToken() == nil {
		return 0
	}
	maxOccurrences := elem.GetPatternToken().MaxOccurrence
	if maxOccurrences == -1 {
		maxOccurrences = len(tokens) // finite bound
	}
	if maxOccurrences < 1 {
		maxOccurrences = 1
	}
	maxSkip := 0
	for j := 1; j < maxOccurrences && mm+j < len(tokens)-remainingElems; j++ {
		if m.testAllReadings(tokens, elem, prevElement, mm+j, firstMatchToken, prevSkipNext, bag) {
			maxSkip++
		} else {
			break
		}
	}
	return maxSkip
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

	phraseCtx := m.phraseMatchContext()
	errMessage := FormatMatches(tokens, tokenPositions, firstMatchToken, msg, sugMatches, lang, phraseCtx)
	shortErrMessage := ""
	if shortMsg != "" {
		shortErrMessage = FormatMatches(tokens, tokenPositions, firstMatchToken, shortMsg, sugMatches, lang, phraseCtx)
	}
	suggestionsOutMsg := FormatMatches(tokens, tokenPositions, firstMatchToken, sugOut, sugMatchesOut, lang, phraseCtx)

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
	// Java: ruleMatch.setType(rule.getType())
	if pr != nil {
		rm.SetType(pr.GetMatchType())
		if pr.IssueType != "" {
			rm.IssueType = pr.IssueType
		}
	}
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
			for _, e := range ExpandSuggestionTemplate(t, tokens, tokenPositions, firstMatchToken, sugMatches, lang, phraseCtx) {
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

// phraseMatchContext builds FormatMatches phrase length context from the matcher.
func (m *PatternRuleMatcher) phraseMatchContext() PhraseMatchContext {
	if m == nil {
		return PhraseMatchContext{}
	}
	ctx := PhraseMatchContext{UseList: m.UseList}
	if m.Rule != nil && m.Rule.PatternRule != nil {
		ctx.ElementNo = m.Rule.PatternRule.ElementNo
		if !ctx.UseList {
			ctx.UseList = m.Rule.PatternRule.UseList
		}
	}
	return ctx
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
