package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// DisambiguatorAction ports DisambiguationPatternRule.DisambiguatorAction.
type DisambiguatorAction string

const (
	ActionAdd            DisambiguatorAction = "ADD"
	ActionFilter         DisambiguatorAction = "FILTER"
	ActionRemove         DisambiguatorAction = "REMOVE"
	ActionReplace        DisambiguatorAction = "REPLACE"
	ActionUnify          DisambiguatorAction = "UNIFY"
	ActionImmunize       DisambiguatorAction = "IMMUNIZE"
	ActionIgnoreSpelling DisambiguatorAction = "IGNORE_SPELLING"
	ActionFilterAll      DisambiguatorAction = "FILTERALL"
	ActionAddChunk       DisambiguatorAction = "ADDCHUNK"
)

// DisambiguationPatternRule ports
// org.languagetool.tagging.disambiguation.rules.DisambiguationPatternRule.
type DisambiguationPatternRule struct {
	*patterns.AbstractTokenBasedRule
	DisambiguatedPOS  string
	MatchElement      *patterns.Match
	Action            DisambiguatorAction
	NewTokenReadings  []*languagetool.AnalyzedToken
	Examples          []DisambiguatedExample
	UntouchedExamples []string
	// AntiPatterns ports Java DisambiguationPatternRule anti-patterns
	// (keepByDisambig overlap suppression).
	AntiPatterns []*patterns.AbstractTokenBasedRule
	// UnifyFeatures are Java <unify><feature id="…"/> names (soft UNIFY).
	UnifyFeatures []string
	// UnifierConfig is the language-level equivalence table (shared).
	UnifierConfig *patterns.UnifierConfiguration
}

func NewDisambiguationPatternRule(
	id, description, languageCode string,
	patternTokens []*patterns.PatternToken,
	disambiguatedPOS string,
	posSelect *patterns.Match,
	action DisambiguatorAction,
) *DisambiguationPatternRule {
	// Java allows ADD/REMOVE/IMMUNIZE/… with only <wd> list (set after construct).
	// Empty postag + empty action defaults to REPLACE in the loader.
	if disambiguatedPOS == "" && posSelect == nil &&
		action != ActionUnify && action != ActionAdd && action != ActionRemove &&
		action != ActionImmunize && action != ActionReplace && action != ActionFilter &&
		action != ActionFilterAll && action != ActionIgnoreSpelling && action != ActionAddChunk &&
		action != "" {
		// Unknown action without POS — skip via panic only for programming errors.
		// Loader should not invent POS; callers must pass a known action.
		panic("disambiguated POS cannot be null with posSelect == null and " + string(action))
	}
	base := patterns.NewAbstractTokenBasedRule(id, description, languageCode, patternTokens)
	base.Message = ""
	// Java DisambiguationPatternRule: super(..., true) — getUnified always true
	// so testUnification stores getFinalUnified for UNIFY actions.
	base.GetUnified = true
	return &DisambiguationPatternRule{
		AbstractTokenBasedRule: base,
		DisambiguatedPOS:       disambiguatedPOS,
		MatchElement:           posSelect,
		Action:                 action,
	}
}

func (r *DisambiguationPatternRule) SetNewInterpretations(readings []*languagetool.AnalyzedToken) {
	r.NewTokenReadings = append([]*languagetool.AnalyzedToken(nil), readings...)
}

func (r *DisambiguationPatternRule) SetExamples(examples []DisambiguatedExample) {
	r.Examples = append([]DisambiguatedExample(nil), examples...)
}

func (r *DisambiguationPatternRule) GetExamples() []DisambiguatedExample { return r.Examples }

func (r *DisambiguationPatternRule) SetUntouchedExamples(ex []string) {
	r.UntouchedExamples = append([]string(nil), ex...)
}

func (r *DisambiguationPatternRule) GetUntouchedExamples() []string { return r.UntouchedExamples }

// SetAntiPatterns ports AbstractPatternRule.setAntiPatterns (appends, Java addAll).
func (r *DisambiguationPatternRule) SetAntiPatterns(aps []*patterns.AbstractTokenBasedRule) {
	if r == nil || len(aps) == 0 {
		return
	}
	r.AntiPatterns = append(r.AntiPatterns, aps...)
}

// Replace applies the disambiguation pattern to the sentence.
// Ports DisambiguationPatternRuleReplacer.replace + executeAction span math.
func (r *DisambiguationPatternRule) Replace(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil || r == nil {
		return sentence
	}
	if r.CanBeIgnoredFor(sentence) {
		return sentence
	}
	base := r.AbstractTokenBasedRule
	if base == nil {
		return sentence
	}
	// Java DisambiguationPatternRuleReplacer: doMatch consumer + executeAction.
	// DoMatch provides tokenPositions (includes skip gaps for startPositionCorrection).
	nws := sentence.GetTokensWithoutWhitespace()
	changed := false
	perf := patterns.NewAbstractPatternRulePerformer(base, nil)
	perf.DoMatch(sentence, func(positions []int, firstMatch, lastMatch, firstMark, lastMark int, unifiedTokens []*languagetool.AnalyzedTokenReadings) {
		if firstMatch < 0 || firstMatch >= len(nws) || lastMatch < 0 || lastMatch >= len(nws) {
			return
		}
		// Java keepDespiteFilter before keepByDisambig
		if !r.keepDespiteFilter(nws, positions, firstMatch, lastMatch) {
			return
		}
		// Java keepByDisambig on full pattern char offsets (firstMatch…lastMatch)
		fromPos := nws[firstMatch].GetStartPos()
		toPos := nws[lastMatch].GetEndPos()
		if !r.keepByDisambig(sentence, fromPos, toPos) {
			return
		}
		if lastMark < 0 {
			lastMark = lastMatch
		}
		// Java executeAction(firstMatchToken, lastMarkerMatchToken, tokenPositions)
		first, last := r.actionSpan(firstMatch, lastMark, nws, positions)
		if first < 0 || last < first {
			return
		}
		r.applyAction(nws, first, last, firstMatch, positions, unifiedTokens)
		changed = true
	})
	if !changed {
		return sentence
	}
	return languagetool.NewAnalyzedSentence(sentence.GetTokens())
}

// indexByCharSpan finds first/last non-whitespace token indices for [fromPos,toPos].
func indexByCharSpan(nws []*languagetool.AnalyzedTokenReadings, fromPos, toPos int) (first, last int) {
	first, last = -1, -1
	for i, t := range nws {
		if t == nil {
			continue
		}
		if t.GetStartPos() == fromPos {
			first = i
		}
		if t.GetEndPos() == toPos {
			last = i
		}
	}
	if first >= 0 && last >= 0 {
		return first, last
	}
	// range fallback
	for i, t := range nws {
		if t == nil {
			continue
		}
		if t.GetStartPos() >= fromPos && (first < 0 || t.GetStartPos() < nws[first].GetStartPos()) {
			first = i
		}
		if t.GetStartPos() < toPos {
			last = i
		}
	}
	return first, last
}

// actionSpan ports executeAction startPositionCorrection using tokenPositions
// when available; falls back to unit-position sum when positions are nil.
// lastMarker is Java lastMarkerMatchToken passed as lastMatchToken.
func (r *DisambiguationPatternRule) actionSpan(firstMatchToken, lastMarker int, nws []*languagetool.AnalyzedTokenReadings, tokenPositions []int) (first, last int) {
	startCorr := 0
	if r.PatternRule != nil {
		startCorr = r.StartPositionCorrection
		if startCorr == 0 && r.EndPositionCorrection == 0 && len(r.Tokens) > 0 {
			startCorr, _ = patterns.PositionCorrectionsFromTokens(r.Tokens)
		}
	}
	// Java: if startPositionCorrection > 0 { correctedStPos--; for l<=corr: += positions[l] }
	from := firstMatchToken
	if startCorr > 0 {
		corrected := -1
		if len(tokenPositions) > 0 {
			lim := startCorr
			if lim > len(tokenPositions)-1 {
				lim = len(tokenPositions) - 1
			}
			for l := 0; l <= lim; l++ {
				corrected += tokenPositions[l]
			}
		} else {
			// unit positions: sum(1 for 0..startCorr) - 1 = startCorr
			corrected = startCorr
		}
		from = firstMatchToken + corrected
	}
	if from < 0 {
		from = 0
	}
	if from >= len(nws) {
		return -1, -1
	}
	to := lastMarker
	if to < from {
		to = from
	}
	if to >= len(nws) {
		to = len(nws) - 1
	}
	return from, to
}

// keepDespiteFilter ports DisambiguationPatternRuleReplacer.keepDespiteFilter.
// When a RuleFilter is set, only apply the action if filter.matches (AcceptRuleMatch ≠ nil).
func (r *DisambiguationPatternRule) keepDespiteFilter(
	tokens []*languagetool.AnalyzedTokenReadings,
	tokenPositions []int,
	firstMatchToken, lastMatchToken int,
) bool {
	if r == nil || r.PatternRule == nil || r.Filter == nil {
		return true
	}
	if firstMatchToken < 0 || lastMatchToken < firstMatchToken || lastMatchToken >= len(tokens) {
		return true
	}
	eval := patterns.NewRuleFilterEvaluator(r.Filter)
	// Java: RuleMatch fakeMatch = new RuleMatch(new FakeRule(), null, 0, 1, "(internal rule)");
	fake := rules.NewRuleMatch(nil, nil, 0, 1, "(internal rule)")
	patternTokens := tokens[firstMatchToken : lastMatchToken+1]
	out := eval.RunFilter(r.FilterArgs, fake, patternTokens, firstMatchToken, tokenPositions)
	return out != nil
}

// keepByDisambig ports DisambiguationPatternRuleReplacer.keepByDisambig:
// false when any anti-pattern match overlaps [fromPos,toPos].
func (r *DisambiguationPatternRule) keepByDisambig(sentence *languagetool.AnalyzedSentence, fromPos, toPos int) bool {
	if r == nil || len(r.AntiPatterns) == 0 {
		return true
	}
	for _, ap := range r.AntiPatterns {
		if ap == nil {
			continue
		}
		if ap.CanBeIgnoredFor(sentence) {
			continue
		}
		// Java: new PatternRuleMatcher(antiPattern, false) — non-pre-disambig.
		// Soft: non-strict POS so open-class anti patterns still fire without a tagger.
		am := patterns.NewPatternRuleMatcher(ap)
		antiMatches, err := am.Match(sentence)
		if err != nil || len(antiMatches) == 0 {
			continue
		}
		for _, dm := range antiMatches {
			if dm == nil {
				continue
			}
			// left overlap of rule match start, right of end, or anti inside rule match
			if (dm.FromPos <= fromPos && dm.ToPos >= fromPos) ||
				(dm.FromPos <= toPos && dm.ToPos >= toPos) ||
				(dm.FromPos >= fromPos && dm.ToPos <= toPos) {
				return false
			}
		}
	}
	return true
}

// applyFilterAll ports DisambiguationPatternRuleReplacer case FILTERALL.
func (r *DisambiguationPatternRule) applyFilterAll(
	nws []*languagetool.AnalyzedTokenReadings,
	firstMatchToken int,
	tokenPositions []int,
) {
	if r == nil || len(r.Tokens) == 0 || firstMatchToken < 0 {
		return
	}
	startCorr := 0
	endCorr := 0
	if r.PatternRule != nil {
		startCorr = r.StartPositionCorrection
		endCorr = r.EndPositionCorrection
		if startCorr == 0 && endCorr == 0 {
			startCorr, endCorr = patterns.PositionCorrectionsFromTokens(r.Tokens)
		}
	}
	// matchingTokens = count of non-zero positions
	matchingTokens := 0
	for _, p := range tokenPositions {
		if p != 0 {
			matchingTokens++
		}
	}
	correctedStPos := 0
	startPositionCorrection := startCorr
	endPositionCorrection := endCorr
	if startPositionCorrection > 0 {
		correctedStPos = -1
		lim := startPositionCorrection
		if lim > len(tokenPositions)-1 {
			lim = len(tokenPositions) - 1
		}
		for l := 0; l <= lim && l < len(tokenPositions); l++ {
			correctedStPos += tokenPositions[l]
		}
		// adjust when optional elements (position 0) appear before start correction
		w := startPositionCorrection
		for j := 0; j <= w && j < len(tokenPositions); j++ {
			if tokenPositions[j] == 0 {
				startPositionCorrection--
			}
		}
	}
	if endPositionCorrection < 0 {
		for d := startPositionCorrection; d < len(tokenPositions); d++ {
			if tokenPositions[d] == 0 {
				endPositionCorrection++
			}
		}
	}
	matchingTokensWithCorrection := matchingTokens
	// lastMatchToken not available here; actionSpan already bounded the marker span.
	// Use positions-based count for the corrected span length.
	spanCount := matchingTokensWithCorrection - startPositionCorrection + endPositionCorrection
	if spanCount <= 0 {
		// Fallback: 1:1 marked tokens (legacy path when positions are empty)
		var marked []*patterns.PatternToken
		for _, pt := range r.Tokens {
			if pt != nil && pt.InsideMarker {
				marked = append(marked, pt)
			}
		}
		if len(marked) == 0 || firstMatchToken+len(marked) > len(nws) {
			return
		}
		for j, pt := range marked {
			pos := firstMatchToken + correctedStPos + j
			if pos < 0 || pos >= len(nws) || nws[pos] == nil || pt == nil || pt.Pos == nil || pt.Pos.PosTag == "" {
				continue
			}
			tmpMatch := patterns.NewMatch(pt.Pos.PosTag, "", true, "", "", patterns.CaseNone, false, false, patterns.IncludeNone)
			ms := tmpMatch.CreateStateWithSynth(nil, nws[pos])
			if filtered := ms.FilterReadings(); filtered != nil {
				nws[pos].ReplaceReadings(filtered.GetReadings(), r.ID)
			}
		}
		return
	}
	for i := 0; i < spanCount; i++ {
		position := firstMatchToken + correctedStPos + i
		if position < 0 || position >= len(nws) || nws[position] == nil {
			continue
		}
		idx := i + startPositionCorrection
		var pToken *patterns.PatternToken
		if idx >= 0 && idx < len(tokenPositions) && tokenPositions[idx] > 0 && idx < len(r.Tokens) {
			pToken = r.Tokens[idx]
		} else {
			k := 1
			for idx+k < len(r.Tokens)+endPositionCorrection &&
				idx+k < len(tokenPositions) &&
				tokenPositions[idx+k] == 0 {
				k++
			}
			if idx+k >= 0 && idx+k < len(r.Tokens) {
				pToken = r.Tokens[idx+k]
			}
		}
		if pToken == nil || pToken.Pos == nil || pToken.Pos.PosTag == "" {
			continue
		}
		// Java: new Match(pToken.getPOStag(), null, true, pToken.getPOStag(), null, …)
		tmpMatch := patterns.NewMatch(pToken.Pos.PosTag, "", true, "", "", patterns.CaseNone, false, false, patterns.IncludeNone)
		ms := tmpMatch.CreateStateWithSynth(nil, nws[position])
		if filtered := ms.FilterReadings(); filtered != nil {
			nws[position].ReplaceReadings(filtered.GetReadings(), r.ID)
		}
	}
}

func (r *DisambiguationPatternRule) applyAction(
	nws []*languagetool.AnalyzedTokenReadings,
	first, last, firstMatchToken int,
	tokenPositions []int,
	unifiedTokens []*languagetool.AnalyzedTokenReadings,
) {
	switch r.Action {
	case ActionImmunize:
		// Java: immunize(rule.getXmlLineNumber())
		line := 0
		if r.PatternRule != nil {
			line = r.LineNumber
			if line < 0 {
				line = 0
			}
		}
		for i := first; i <= last && i < len(nws); i++ {
			if nws[i] != nil {
				nws[i].Immunize(line)
			}
		}
	case ActionRemove:
		// Java REMOVE (DisambiguationPatternRuleReplacer):
		// 1) <wd> list: only when length equals marker-span token count; each
		//    wd removes matching readings via AnalyzedToken.matches (partial).
		// 2) else disambiguatedPOS: negative POS *regex* filter on fromPos only.
		if len(r.NewTokenReadings) > 0 {
			span := 0
			if last >= first && first >= 0 {
				span = last - first + 1
			}
			// Java: newTokenReadings.length == matchingTokensWithCorrection - …
			if span == 0 || len(r.NewTokenReadings) != span {
				return
			}
			for i := first; i <= last && i < len(nws); i++ {
				rel := i - first
				if r.NewTokenReadings[rel] == nil || nws[i] == nil {
					continue
				}
				nws[i].RemoveReading(r.NewTokenReadings[rel], r.ID)
			}
			return
		}
		if r.DisambiguatedPOS == "" || first < 0 || first >= len(nws) || nws[first] == nil {
			return
		}
		re, err := regexp.Compile("^(?:" + r.DisambiguatedPOS + ")$")
		if err != nil {
			return
		}
		for _, reading := range append([]*languagetool.AnalyzedToken(nil), nws[first].GetReadings()...) {
			if reading.GetPOSTag() != nil && re.MatchString(*reading.GetPOSTag()) {
				nws[first].RemoveReading(reading, r.ID)
			}
		}
	case ActionAdd:
		// Java ADD (DisambiguationPatternRuleReplacer): <wd> list only when
		// length equals marker-span count; empty wd surface uses matched token.
		// Bare postag adds that POS to each matched token without a wd list.
		if len(r.NewTokenReadings) > 0 {
			span := 0
			if last >= first && first >= 0 {
				span = last - first + 1
			}
			if span == 0 || len(r.NewTokenReadings) != span {
				return
			}
			for i := first; i <= last && i < len(nws); i++ {
				rel := i - first
				if nws[i] == nil || r.NewTokenReadings[rel] == nil {
					continue
				}
				base := r.NewTokenReadings[rel]
				surface := nws[i].GetToken()
				if base.GetToken() != "" {
					surface = base.GetToken()
				}
				tok := languagetool.NewAnalyzedToken(surface, base.GetPOSTag(), base.GetLemma())
				nws[i].AddReading(tok, r.ID)
			}
			return
		}
		if r.DisambiguatedPOS == "" {
			return
		}
		pos := r.DisambiguatedPOS
		for i := first; i <= last && i < len(nws); i++ {
			if nws[i] == nil {
				continue
			}
			surface := nws[i].GetToken()
			nws[i].AddReading(languagetool.NewAnalyzedToken(surface, &pos, nil), r.ID)
		}
	case ActionReplace:
		// Java REPLACE with <wd> list: only when length equals marker-span count
		// (DisambiguationPatternRuleReplacer: newTokenReadings.length == …).
		// Java REPLACE with only postag: replace *fromPos* only
		// (whTokens[fromPos] = new …). first is matcher marker start (fromPos).
		// Java REPLACE with matchElement: MatchState.filterReadings().
		if len(r.NewTokenReadings) > 0 {
			span := 0
			if last >= first && first >= 0 {
				span = last - first + 1
			}
			if span == 0 || len(r.NewTokenReadings) != span {
				return
			}
			for i := first; i <= last && i < len(nws); i++ {
				rel := i - first
				if r.NewTokenReadings[rel] == nil || nws[i] == nil {
					continue
				}
				tok := r.NewTokenReadings[rel]
				surface := nws[i].GetToken()
				if tok.GetToken() != "" {
					surface = tok.GetToken()
				}
				pos := tok.GetPOSTag()
				lemma := tok.GetLemma()
				if lemma == nil {
					lemma = &surface
				}
				newTok := languagetool.NewAnalyzedToken(surface, pos, lemma)
				nws[i].ReplaceReadings([]*languagetool.AnalyzedToken{newTok}, r.ID)
			}
			return
		}
		if first < 0 || first >= len(nws) || nws[first] == nil {
			return
		}
		if r.MatchElement != nil {
			// Java: matchElement.createState(synth, whTokens[fromPos]).filterReadings()
			// Mutate in place — nws aliases sentence token pointers.
			ms := r.MatchElement.CreateStateWithSynth(nil, nws[first])
			filtered := ms.FilterReadings()
			if filtered != nil {
				nws[first].ReplaceReadings(filtered.GetReadings(), r.ID)
			}
			return
		}
		if r.DisambiguatedPOS == "" {
			return
		}
		surface := nws[first].GetToken()
		lemma := surface
		for _, reading := range nws[first].GetReadings() {
			if reading.GetPOSTag() != nil && *reading.GetPOSTag() == r.DisambiguatedPOS && reading.GetLemma() != nil {
				lemma = *reading.GetLemma()
				break
			}
		}
		if lemma == surface {
			if at := nws[first].GetAnalyzedToken(0); at != nil && at.GetLemma() != nil && *at.GetLemma() != "" {
				lemma = *at.GetLemma()
			}
		}
		pos := r.DisambiguatedPOS
		newTok := languagetool.NewAnalyzedToken(surface, &pos, &lemma)
		nws[first].ReplaceReadings([]*languagetool.AnalyzedToken{newTok}, r.ID)
	case ActionFilter:
		// Java FILTER: when matchElement==null, build Match(disambiguatedPOS, postagRegexp)
		// and apply filterReadings only if some reading already matches the POS regex.
		// When matchElement!=null, fall through as REPLACE (Java case FILTER fallthrough).
		if r.MatchElement != nil {
			if first < 0 || first >= len(nws) || nws[first] == nil {
				return
			}
			ms := r.MatchElement.CreateStateWithSynth(nil, nws[first])
			filtered := ms.FilterReadings()
			if filtered != nil {
				nws[first].ReplaceReadings(filtered.GetReadings(), r.ID)
			}
			return
		}
		if r.DisambiguatedPOS == "" || first < 0 || first >= len(nws) || nws[first] == nil {
			return
		}
		// Java: Match(disambiguatedPOS, null, true, disambiguatedPOS, null, …)
		tmpMatch := patterns.NewMatch(r.DisambiguatedPOS, "", true, "", "", patterns.CaseNone, false, false, patterns.IncludeNone)
		// newPOSmatches gate (RE2 or lookaround POS engines)
		if !tmpMatch.HasPosRegexp() {
			return
		}
		newPOSmatches := false
		for _, reading := range nws[first].GetReadings() {
			if reading == nil || reading.HasNoTag() {
				continue
			}
			// Java String.matches / Matcher.matches — full POS string
			if pt := reading.GetPOSTag(); pt != nil && tmpMatch.PosFullMatch(*pt) {
				newPOSmatches = true
				break
			}
		}
		if !newPOSmatches {
			return
		}
		ms := tmpMatch.CreateStateWithSynth(nil, nws[first])
		filtered := ms.FilterReadings()
		if filtered != nil {
			nws[first].ReplaceReadings(filtered.GetReadings(), r.ID)
		}
	case ActionFilterAll:
		// Java FILTERALL (DisambiguationPatternRuleReplacer): map each matched
		// token in the corrected span to a pattern token via tokenPositions
		// (skip zero-width optional elements), then MatchState.filterReadings.
		r.applyFilterAll(nws, firstMatchToken, tokenPositions)
	case ActionIgnoreSpelling:
		for i := first; i <= last && i < len(nws); i++ {
			if nws[i] != nil {
				nws[i].IgnoreSpelling()
			}
		}
	case ActionAddChunk:
		// Java ADDCHUNK: each <wd pos="…"/> is a ChunkTag appended to the
		// matching marker-span token (length must equal span). Soft also
		// accepts a single DisambiguatedPOS when no <wd> list is present.
		if len(r.NewTokenReadings) > 0 {
			span := 0
			if last >= first && first >= 0 {
				span = last - first + 1
			}
			if span == 0 || len(r.NewTokenReadings) != span {
				return
			}
			for i := first; i <= last && i < len(nws); i++ {
				rel := i - first
				if nws[i] == nil || r.NewTokenReadings[rel] == nil {
					continue
				}
				pos := r.NewTokenReadings[rel].GetPOSTag()
				if pos == nil || *pos == "" {
					continue
				}
				tags := append([]string(nil), nws[i].GetChunkTags()...)
				dup := false
				for _, t := range tags {
					if t == *pos {
						dup = true
						break
					}
				}
				if !dup {
					tags = append(tags, *pos)
				}
				nws[i].SetChunkTags(tags)
			}
			return
		}
		if r.DisambiguatedPOS == "" {
			return
		}
		for i := first; i <= last && i < len(nws); i++ {
			if nws[i] == nil {
				continue
			}
			tags := append([]string(nil), nws[i].GetChunkTags()...)
			dup := false
			for _, t := range tags {
				if t == r.DisambiguatedPOS {
					dup = true
					break
				}
			}
			if !dup {
				tags = append(tags, r.DisambiguatedPOS)
			}
			nws[i].SetChunkTags(tags)
		}
	case ActionUnify:
		// Java UNIFY: apply match-time unifiedTokens (getUnified=true) to the
		// corrected marker span. Length must equal span token count.
		r.applyUnify(nws, first, last, firstMatchToken, tokenPositions, unifiedTokens)
	default:
		_ = fmt.Sprintf
	}
}

// applyUnify ports DisambiguationPatternRuleReplacer case UNIFY.
// Prefer match-time unifiedTokens from AbstractPatternRulePerformer; fall back
// to re-running the unifier over the span when tokens were not captured.
func (r *DisambiguationPatternRule) applyUnify(
	nws []*languagetool.AnalyzedTokenReadings,
	first, last, firstMatchToken int,
	tokenPositions []int,
	unifiedTokens []*languagetool.AnalyzedTokenReadings,
) {
	if r == nil || first < 0 || last < first || last >= len(nws) {
		return
	}
	// Java: if (unifiedTokens != null && length == matchingTokensWithCorrection - start + end)
	if len(unifiedTokens) > 0 {
		span := last - first + 1
		if len(unifiedTokens) != span {
			// Still try when lengths match after correction math fails open-closed.
			// Java only applies when lengths equal.
			return
		}
		if nws[last] != nil && nws[last].IsSentenceEnd() && unifiedTokens[len(unifiedTokens)-1] != nil {
			unifiedTokens[len(unifiedTokens)-1].SetSentEnd()
		}
		for i := 0; i < len(unifiedTokens); i++ {
			pos := first + i
			if pos >= len(nws) || nws[pos] == nil || unifiedTokens[i] == nil {
				continue
			}
			nws[pos].ReplaceReadings(append([]*languagetool.AnalyzedToken(nil), unifiedTokens[i].GetReadings()...), r.ID)
		}
		return
	}
	// Fallback when getUnified did not capture tokens (missing UniFeatures on pattern).
	if r.UnifierConfig == nil || len(r.UnifyFeatures) == 0 {
		return
	}
	uFeatures := make(map[string][]string, len(r.UnifyFeatures))
	for _, f := range r.UnifyFeatures {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		uFeatures[f] = nil
	}
	if len(uFeatures) == 0 {
		return
	}
	uni := r.UnifierConfig.CreateUnifier()
	for i := first; i <= last; i++ {
		if nws[i] == nil {
			return
		}
		readings := nws[i].GetReadings()
		if len(readings) == 0 {
			return
		}
		for j, rd := range readings {
			if rd == nil {
				continue
			}
			lastReading := j == len(readings)-1
			uni.IsUnifiedMatched(rd, uFeatures, lastReading, true)
		}
	}
	unified := uni.GetFinalUnified()
	span := last - first + 1
	if unified == nil || len(unified) != span {
		return
	}
	for j, i := 0, first; i <= last; i, j = i+1, j+1 {
		if unified[j] == nil || nws[i] == nil {
			continue
		}
		nws[i].ReplaceReadings(append([]*languagetool.AnalyzedToken(nil), unified[j].GetReadings()...), r.ID)
	}
}


