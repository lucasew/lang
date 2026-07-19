package pt

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// RegularIrregularParticipleFilter ports org.languagetool.rules.pt.RegularIrregularParticipleFilter.
// Synthesize produces participle candidates for a VMP reading (Java: PortugueseSynthesizer).
type RegularIrregularParticipleFilter struct {
	// Synthesize(lemma, desiredPostag) → surface forms (regex postag). Nil → fail-closed.
	Synthesize func(lemma, desiredPostag string) []string
}

func NewRegularIrregularParticipleFilter() *RegularIrregularParticipleFilter {
	return &RegularIrregularParticipleFilter{}
}

// IsRegular reports regular PT participle endings (do/dos/da/das).
func IsRegularParticiple(p string) bool {
	lp := strings.ToLower(p)
	return strings.HasSuffix(lp, "do") || strings.HasSuffix(lp, "dos") ||
		strings.HasSuffix(lp, "da") || strings.HasSuffix(lp, "das")
}

// AcceptRuleMatch ports RegularIrregularParticipleFilter.acceptRuleMatch.
// Args: direction — "RegularToIrregular" or "IrregularToRegular".
func (f *RegularIrregularParticipleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	direction, ok := arguments["direction"]
	if !ok {
		panic("Missing key 'direction'")
	}
	// Find pattern token aligned with match start (Java: getStartPos() == match.getFromPos()).
	var atr *languagetool.AnalyzedTokenReadings
	for _, tok := range patternTokens {
		if tok == nil {
			continue
		}
		if tok.GetStartPos() == match.GetFromPos() {
			atr = tok
			break
		}
	}
	if atr == nil {
		// Fallback: single-token pattern matches often use only the marked token.
		if len(patternTokens) == 1 {
			atr = patternTokens[0]
		} else {
			return nil
		}
	}
	if !atr.HasPosTagStartingWith("VMP") {
		return nil
	}
	var selectedLemma, desiredPostag string
	for _, at := range atr.GetReadings() {
		if at == nil || at.GetPOSTag() == nil {
			continue
		}
		pos := *at.GetPOSTag()
		if strings.HasPrefix(pos, "VMP") {
			if at.GetLemma() != nil {
				selectedLemma = *at.GetLemma()
			} else {
				selectedLemma = at.GetToken()
			}
			desiredPostag = pos
		}
	}
	if desiredPostag == "" || f.Synthesize == nil {
		return nil
	}
	// Java: if ends with C → [MC]; else last char → [last C]
	if strings.HasSuffix(desiredPostag, "C") {
		desiredPostag = desiredPostag[:len(desiredPostag)-1] + "[MC]"
	} else {
		last := desiredPostag[len(desiredPostag)-1:]
		desiredPostag = desiredPostag[:len(desiredPostag)-1] + "[" + last + "C]"
	}
	participles := f.Synthesize(selectedLemma, desiredPostag)
	template := ""
	if reps := match.GetSuggestedReplacements(); len(reps) > 0 {
		template = reps[0]
	}
	replacement := f.Suggest(direction, atr.GetToken(), participles, template)
	if replacement == "" {
		return nil
	}
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.ShortMessage
	out.SetSuggestedReplacement(replacement)
	return out
}

// Suggest picks a regular/irregular alternate from synthesizer candidates.
// direction is "RegularToIrregular" or "IrregularToRegular".
// token is the matched participle form; candidates are synthesised forms.
// template may contain {suggestion}/{Suggestion}/{SUGGESTION}.
func (f *RegularIrregularParticipleFilter) Suggest(direction, token string, candidates []string, template string) string {
	if len(candidates) < 2 {
		return ""
	}
	var replacement string
	dir := strings.ToLower(direction)
	if dir == strings.ToLower("RegularToIrregular") && IsRegularParticiple(token) {
		if !IsRegularParticiple(candidates[0]) {
			replacement = candidates[0]
		} else if !IsRegularParticiple(candidates[1]) {
			replacement = candidates[1]
		}
	} else if dir == strings.ToLower("IrregularToRegular") && !IsRegularParticiple(token) {
		if IsRegularParticiple(candidates[0]) {
			replacement = candidates[0]
		} else if IsRegularParticiple(candidates[1]) {
			replacement = candidates[1]
		}
	}
	if replacement == "" {
		return ""
	}
	if template == "" {
		return replacement
	}
	s := strings.ReplaceAll(template, "{suggestion}", replacement)
	s = strings.ReplaceAll(s, "{Suggestion}", tools.UppercaseFirstChar(replacement))
	s = strings.ReplaceAll(s, "{SUGGESTION}", strings.ToUpper(replacement))
	return s
}
