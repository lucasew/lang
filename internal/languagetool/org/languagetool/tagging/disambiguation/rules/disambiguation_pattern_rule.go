package rules

import (
	"fmt"
	"regexp"

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
}

func NewDisambiguationPatternRule(
	id, description, languageCode string,
	patternTokens []*patterns.PatternToken,
	disambiguatedPOS string,
	posSelect *patterns.Match,
	action DisambiguatorAction,
) *DisambiguationPatternRule {
	if disambiguatedPOS == "" && posSelect == nil &&
		action != ActionUnify && action != ActionAdd && action != ActionRemove &&
		action != ActionImmunize && action != ActionReplace && action != ActionFilterAll &&
		action != ActionIgnoreSpelling && action != ActionAddChunk {
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

func (r *DisambiguationPatternRule) applyAction(nws []*languagetool.AnalyzedTokenReadings, first, last int) {
	switch r.Action {
	case ActionImmunize:
		for i := first; i <= last && i < len(nws); i++ {
			nws[i].Immunize(0)
		}
	case ActionRemove:
		// Java REMOVE with disambiguatedPOS: negative filtering via POS *regex*
		// on fromPos (first matched token), not exact string equality.
		if r.DisambiguatedPOS == "" {
			if len(r.NewTokenReadings) > 0 {
				for i := first; i <= last && i < len(nws); i++ {
					rel := i - first
					if rel < len(r.NewTokenReadings) && r.NewTokenReadings[rel] != nil {
						nws[i].RemoveReading(r.NewTokenReadings[rel], r.ID)
					}
				}
			}
			return
		}
		re, err := regexp.Compile("^(?:" + r.DisambiguatedPOS + ")$")
		if err != nil || first < 0 || first >= len(nws) || nws[first] == nil {
			return
		}
		for _, reading := range append([]*languagetool.AnalyzedToken(nil), nws[first].GetReadings()...) {
			if reading.GetPOSTag() != nil && re.MatchString(*reading.GetPOSTag()) {
				nws[first].RemoveReading(reading, r.ID)
			}
		}
	case ActionAdd:
		// Java ADD: empty <wd> surface uses the matched token string.
		for i := first; i <= last && i < len(nws); i++ {
			if nws[i] == nil {
				continue
			}
			rel := i - first
			var tok *languagetool.AnalyzedToken
			if rel < len(r.NewTokenReadings) && r.NewTokenReadings[rel] != nil {
				base := r.NewTokenReadings[rel]
				surface := nws[i].GetToken()
				if base.GetToken() != "" {
					surface = base.GetToken()
				}
				tok = languagetool.NewAnalyzedToken(surface, base.GetPOSTag(), base.GetLemma())
			} else if r.DisambiguatedPOS != "" {
				pos := r.DisambiguatedPOS
				surface := nws[i].GetToken()
				tok = languagetool.NewAnalyzedToken(surface, &pos, nil)
			} else if len(r.NewTokenReadings) == 1 && r.NewTokenReadings[0] != nil {
				// Single <wd/> applied to every matched token (UNKNOWN_PCT style).
				base := r.NewTokenReadings[0]
				surface := nws[i].GetToken()
				if base.GetToken() != "" {
					surface = base.GetToken()
				}
				tok = languagetool.NewAnalyzedToken(surface, base.GetPOSTag(), base.GetLemma())
			}
			if tok != nil {
				nws[i].AddReading(tok, r.ID)
			}
		}
	case ActionReplace:
		// Java REPLACE with <wd> list: one reading per matched (marker) token.
		// Java REPLACE with only postag: replace *fromPos* only
		// (DisambiguationPatternRuleReplacer: whTokens[fromPos] = new …).
		// first/last already come from the matcher marker span (firstMark/lastMark),
		// so first == Java fromPos after startPositionCorrection.
		if len(r.NewTokenReadings) > 0 {
			for i := first; i <= last && i < len(nws); i++ {
				rel := i - first
				if rel >= len(r.NewTokenReadings) || r.NewTokenReadings[rel] == nil || nws[i] == nil {
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
	case ActionFilter, ActionFilterAll:
		// Java FILTER (DisambiguationPatternRuleReplacer case FILTER): POS is a
		// regex; apply only when some reading matches (newPOSmatches); keep
		// readings whose POS matches. Target is *fromPos only* (first of the
		// matcher marker span). FILTERALL walks every token in the span.
		if r.DisambiguatedPOS == "" {
			return
		}
		re, err := regexp.Compile("^(?:" + r.DisambiguatedPOS + ")$")
		if err != nil {
			return
		}
		var targets []int
		if r.Action == ActionFilterAll {
			for i := first; i <= last && i < len(nws); i++ {
				targets = append(targets, i)
			}
		} else if first >= 0 {
			// Java: whTokens[fromPos] only — first is already marker-start.
			targets = []int{first}
		}
		for _, i := range targets {
			if i < 0 || i >= len(nws) || nws[i] == nil {
				continue
			}
			// only apply when at least one reading matches (Java newPOSmatches)
			any := false
			for _, reading := range nws[i].GetReadings() {
				if reading.GetPOSTag() != nil && re.MatchString(*reading.GetPOSTag()) {
					any = true
					break
				}
			}
			if !any {
				continue
			}
			for _, reading := range append([]*languagetool.AnalyzedToken(nil), nws[i].GetReadings()...) {
				if reading.GetPOSTag() == nil || !re.MatchString(*reading.GetPOSTag()) {
					nws[i].RemoveReading(reading, r.ID)
				}
			}
		}
	case ActionIgnoreSpelling:
		for i := first; i <= last && i < len(nws); i++ {
			if nws[i] != nil {
				nws[i].IgnoreSpelling()
			}
		}
	case ActionAddChunk:
		// ADDCHUNK: DisambiguatedPOS is the chunk tag to set on matched tokens
		if r.DisambiguatedPOS == "" {
			return
		}
		for i := first; i <= last && i < len(nws); i++ {
			if nws[i] != nil {
				nws[i].SetChunkTags([]string{r.DisambiguatedPOS})
			}
		}
	default:
		// UNIFY deferred
		_ = fmt.Sprintf
	}
}
