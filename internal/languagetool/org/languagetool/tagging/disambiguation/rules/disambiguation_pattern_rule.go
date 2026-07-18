package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
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

// SetAntiPatterns ports AbstractPatternRule.setAntiPatterns for soft disambig.
func (r *DisambiguationPatternRule) SetAntiPatterns(aps []*patterns.AbstractTokenBasedRule) {
	r.AntiPatterns = append([]*patterns.AbstractTokenBasedRule(nil), aps...)
}

// Replace applies the disambiguation pattern to the sentence (simplified actions).
func (r *DisambiguationPatternRule) Replace(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil || r == nil {
		return sentence
	}
	if r.CanBeIgnoredFor(sentence) {
		return sentence
	}
	// Strict POS: untagged tokens are UNKNOWN only (Java with a real tagger).
	// Soft grammar keeps the non-strict matcher so open-class postags soft-match.
	matcher := patterns.NewPatternRuleMatcherStrict(r.AbstractTokenBasedRule)
	matches, err := matcher.Match(sentence)
	if err != nil || len(matches) == 0 {
		return sentence
	}
	// work on a copy of token slice
	tokens := append([]*languagetool.AnalyzedTokenReadings(nil), sentence.GetTokens()...)
	nws := sentence.GetTokensWithoutWhitespace()
	for _, m := range matches {
		// Java keepByDisambig: suppress when an anti-pattern overlaps this match.
		if !r.keepByDisambig(sentence, m.FromPos, m.ToPos) {
			continue
		}
		// map match positions back to non-whitespace tokens by start pos
		first, last := -1, -1
		for i, t := range nws {
			if t.GetStartPos() == m.FromPos {
				first = i
			}
			if t.GetEndPos() == m.ToPos || t.GetStartPos()+len(t.GetToken()) == m.ToPos {
				last = i
			}
		}
		if first < 0 {
			// fallback: find by start pos range
			for i, t := range nws {
				if t.GetStartPos() >= m.FromPos && (first < 0 || t.GetStartPos() < nws[first].GetStartPos()) {
					first = i
				}
				if t.GetStartPos() < m.ToPos {
					last = i
				}
			}
		}
		if first < 0 || last < 0 {
			continue
		}
		r.applyAction(nws, first, last)
	}
	// rebuild sentence from original token order if nws is a view of tokens
	// GetTokensWithoutWhitespace returns subset; mutations on readings objects are shared.
	_ = tokens
	return languagetool.NewAnalyzedSentence(sentence.GetTokens())
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

func (r *DisambiguationPatternRule) applyAction(nws []*languagetool.AnalyzedTokenReadings, first, last int) {
	switch r.Action {
	case ActionImmunize:
		for i := first; i <= last && i < len(nws); i++ {
			nws[i].Immunize(0)
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
		if first < 0 || first >= len(nws) || nws[first] == nil || r.DisambiguatedPOS == "" {
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
		// Java FILTER (DisambiguationPatternRuleReplacer case FILTER): POS is a
		// regex; apply only when some reading matches (newPOSmatches); keep
		// readings whose POS matches. Target is *fromPos only* (first of the
		// matcher marker span — Java whTokens[fromPos]).
		if r.DisambiguatedPOS == "" || first < 0 || first >= len(nws) || nws[first] == nil {
			return
		}
		re, err := regexp.Compile("^(?:" + r.DisambiguatedPOS + ")$")
		if err != nil {
			return
		}
		filterReadingsByPOS(nws[first], re, r.ID)
	case ActionFilterAll:
		// Java FILTERALL: for each *matched* pattern token in the marker span,
		// filter readings by that PatternToken's POS (Match postagRegexp=true).
		// Soft maps 1:1 when marker span length == number of marked pattern
		// tokens (no skip gaps). With skip/min gaps, tokenPositions are not
		// available — leave the span unchanged rather than mis-assign POS.
		var marked []*patterns.PatternToken
		if r.AbstractTokenBasedRule != nil {
			for _, pt := range r.Tokens {
				if pt != nil && pt.InsideMarker {
					marked = append(marked, pt)
				}
			}
		}
		span := 0
		if last >= first && first >= 0 {
			span = last - first + 1
		}
		if span == 0 || len(marked) != span {
			return
		}
		for j, i := 0, first; i <= last && i < len(nws); i, j = i+1, j+1 {
			if nws[i] == nil || j >= len(marked) || marked[j] == nil || marked[j].Pos == nil {
				continue
			}
			posTag := marked[j].Pos.PosTag
			if posTag == "" {
				continue
			}
			// Java always builds Match with postagRegexp=true for FILTERALL.
			re, err := regexp.Compile("^(?:" + posTag + ")$")
			if err != nil {
				continue
			}
			filterReadingsByPOS(nws[i], re, r.ID)
		}
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
		// Java UNIFY: filter matched tokens to readings that share feature
		// combinations (Unifier.getFinalUnified). Soft runs unification after
		// the pattern match (match-time uni is not required for soft extract).
		r.applyUnify(nws, first, last)
	default:
		_ = fmt.Sprintf
	}
}

// applyUnify ports DisambiguationPatternRuleReplacer case UNIFY using the
// language UnifierConfiguration and this rule's UnifyFeatures.
func (r *DisambiguationPatternRule) applyUnify(nws []*languagetool.AnalyzedTokenReadings, first, last int) {
	if r == nil || r.UnifierConfig == nil || len(r.UnifyFeatures) == 0 {
		return
	}
	if first < 0 || last < first || last >= len(nws) {
		return
	}
	uFeatures := make(map[string][]string, len(r.UnifyFeatures))
	for _, f := range r.UnifyFeatures {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		// empty type list → Unifier uses all equivalence types for the feature
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
			// still advance unifier with a dummy empty? leave as fail
			return
		}
		for j, rd := range readings {
			if rd == nil {
				continue
			}
			lastReading := j == len(readings)-1
			// Soft: treat every reading as matched (isMatched=true); Java gates
			// on pattern-token match before isUnified.
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

// filterReadingsByPOS keeps readings whose POS matches re (Java MatchState.filterReadings
// for disambiguation FILTER/FILTERALL). No-op when no reading matches (Java FILTER's
// newPOSmatches gate; FILTERALL empty path leaves the token via soft no-op).
func filterReadingsByPOS(tok *languagetool.AnalyzedTokenReadings, re *regexp.Regexp, ruleID string) {
	if tok == nil || re == nil {
		return
	}
	any := false
	for _, reading := range tok.GetReadings() {
		if reading.GetPOSTag() != nil && re.MatchString(*reading.GetPOSTag()) {
			any = true
			break
		}
	}
	if !any {
		return
	}
	for _, reading := range append([]*languagetool.AnalyzedToken(nil), tok.GetReadings()...) {
		if reading.GetPOSTag() == nil || !re.MatchString(*reading.GetPOSTag()) {
			tok.RemoveReading(reading, ruleID)
		}
	}
}
