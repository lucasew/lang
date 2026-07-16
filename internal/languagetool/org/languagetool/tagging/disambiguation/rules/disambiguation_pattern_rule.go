package rules

import (
	"fmt"

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
	matcher := patterns.NewPatternRuleMatcher(r.AbstractTokenBasedRule)
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
		// remove readings with matching POS if set
		for i := first; i <= last && i < len(nws); i++ {
			if r.DisambiguatedPOS == "" {
				continue
			}
			for _, reading := range nws[i].GetReadings() {
				if reading.GetPOSTag() != nil && *reading.GetPOSTag() == r.DisambiguatedPOS {
					nws[i].RemoveReading(reading, r.ID)
				}
			}
		}
	case ActionAdd, ActionReplace:
		for i := first; i <= last && i < len(nws); i++ {
			rel := i - first
			var tok *languagetool.AnalyzedToken
			if rel < len(r.NewTokenReadings) {
				tok = r.NewTokenReadings[rel]
			} else if r.DisambiguatedPOS != "" {
				pos := r.DisambiguatedPOS
				surface := nws[i].GetToken()
				tok = languagetool.NewAnalyzedToken(surface, &pos, nil)
			}
			if tok == nil {
				continue
			}
			if r.Action == ActionReplace {
				// leave only new reading
				nws[i].LeaveReading(tok)
			} else {
				nws[i].AddReading(tok, r.ID)
			}
		}
	case ActionFilter, ActionFilterAll:
		// keep readings whose POS equals DisambiguatedPOS
		if r.DisambiguatedPOS == "" {
			return
		}
		for i := first; i <= last && i < len(nws); i++ {
			for _, reading := range append([]*languagetool.AnalyzedToken(nil), nws[i].GetReadings()...) {
				if reading.GetPOSTag() == nil || *reading.GetPOSTag() != r.DisambiguatedPOS {
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
