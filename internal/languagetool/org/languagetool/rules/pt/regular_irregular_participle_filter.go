package pt

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	ptsynthesis "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/pt"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// RegularIrregularParticipleFilter ports
// org.languagetool.rules.pt.RegularIrregularParticipleFilter.
//
// Java always holds PortugueseSynthesizer from Portuguese.getInstance().getSynthesizer().
// Without a loaded dict the synthesizer returns no dual participles → fail-closed (null).
type RegularIrregularParticipleFilter struct {
	// synth ports the final PortugueseSynthesizer field.
	// Nil only when explicitly cleared for tests; New* defaults to ptsynthesis.INSTANCE.
	synth synthesis.Synthesizer
}

// defaultParticipleSynth is an optional process-wide override
// (host wires a dict-backed PortugueseSynthesizer, like AdvancedSynthesizerFilter).
var defaultParticipleSynth synthesis.Synthesizer

// WireDefaultParticipleSynthesizer installs the process-wide synthesizer used by
// NewRegularIrregularParticipleFilter (Java: Portuguese.getSynthesizer()).
func WireDefaultParticipleSynthesizer(s synthesis.Synthesizer) {
	defaultParticipleSynth = s
}

// ClearDefaultParticipleSynthesizer clears the process-wide override (tests).
func ClearDefaultParticipleSynthesizer() {
	defaultParticipleSynth = nil
}

// NewRegularIrregularParticipleFilter ports the Java no-arg constructor.
// Defaults to WireDefault if set, else PortugueseSynthesizer.INSTANCE.
func NewRegularIrregularParticipleFilter() *RegularIrregularParticipleFilter {
	f := &RegularIrregularParticipleFilter{}
	if defaultParticipleSynth != nil {
		f.synth = defaultParticipleSynth
	} else {
		// Java field init: (PortugueseSynthesizer) language.getSynthesizer()
		f.synth = ptsynthesis.INSTANCE
	}
	return f
}

// SetSynthesizer overrides the synthesizer (tests / host injection).
func (f *RegularIrregularParticipleFilter) SetSynthesizer(s synthesis.Synthesizer) {
	if f != nil {
		f.synth = s
	}
}

// IsRegularParticiple ports RegularIrregularParticipleFilter.isRegular.
// Regular PT participle endings: do / dos / da / das (case-insensitive).
func IsRegularParticiple(p string) bool {
	lp := strings.ToLower(p)
	return strings.HasSuffix(lp, "do") || strings.HasSuffix(lp, "dos") ||
		strings.HasSuffix(lp, "da") || strings.HasSuffix(lp, "das")
}

// AcceptRuleMatch ports RegularIrregularParticipleFilter.acceptRuleMatch.
// Args: direction — "RegularToIrregular" or "IrregularToRegular" (required).
func (f *RegularIrregularParticipleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	// Java NPE on null this/match; Go nil-receiver safety for host calls.
	if f == nil || match == nil {
		return nil
	}
	// Java: getRequired("direction", arguments) → RuntimeException if missing
	direction := patterns.GetRequired("direction", arguments)

	// Java: find pattern token with token.getStartPos() == match.getFromPos()
	// No single-token fallback — Java leaves atr null and NPEs on hasPosTagStartingWith.
	var atr *languagetool.AnalyzedTokenReadings
	for _, token := range patternTokens {
		if token == nil {
			continue
		}
		if token.GetStartPos() == match.GetFromPos() {
			atr = token
			break
		}
	}
	// Java: atr.hasPosTagStartingWith("VMP") — NPE if atr is null
	if atr == nil {
		// Faithful NPE twin (no invent fallback when start-pos not found).
		panic("null AnalyzedTokenReadings (no pattern token at match.fromPos)")
	}

	var replacement string
	if atr.HasPosTagStartingWith("VMP") {
		var selectedAT *languagetool.AnalyzedToken
		var desiredPostag string
		// Java: last VMP AnalyzedToken wins
		for _, at := range atr.GetReadings() {
			if at == nil {
				continue
			}
			if at.GetPOSTag() != nil && strings.HasPrefix(*at.GetPOSTag(), "VMP") {
				selectedAT = at
				desiredPostag = *at.GetPOSTag()
			}
		}
		// assente SC assentado SM
		if strings.HasSuffix(desiredPostag, "C") {
			desiredPostag = desiredPostag[:len(desiredPostag)-1] + "[MC]"
		} else {
			desiredPostag = desiredPostag[:len(desiredPostag)-1] +
				"[" + desiredPostag[len(desiredPostag)-1:] + "C]"
		}

		// Java: synth.synthesize(selectedAT, desiredPostag, true)
		var participles []string
		if f.synth != nil {
			var err error
			participles, err = f.synth.SynthesizeRE(selectedAT, desiredPostag, true)
			if err != nil {
				// Java: throw new IOException("Cannot synthesize " + selectedAT + e)
				panic("Cannot synthesize " + selectedAT.String() + err.Error())
			}
		}
		// Java: if (participles != null && participles.length > 1)
		if participles != nil && len(participles) > 1 {
			if strings.EqualFold(direction, "RegularToIrregular") && IsRegularParticiple(atr.GetToken()) {
				if !IsRegularParticiple(participles[0]) {
					replacement = participles[0]
				} else if !IsRegularParticiple(participles[1]) {
					replacement = participles[1]
				}
			} else if strings.EqualFold(direction, "IrregularToRegular") && !IsRegularParticiple(atr.GetToken()) {
				if IsRegularParticiple(participles[0]) {
					replacement = participles[0]
				} else if IsRegularParticiple(participles[1]) {
					replacement = participles[1]
				}
			}
			if replacement != "" {
				message := match.GetMessage()
				// Java: new RuleMatch(rule, sentence, from, to, message, shortMessage)
				ruleMatch := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(),
					match.GetToPos(), message)
				ruleMatch.ShortMessage = match.GetShortMessage()
				// Java: ruleMatch.setType(match.getType())
				ruleMatch.SetType(match.GetType())
				// Java: match.getSuggestedReplacements().get(0) — IndexOutOfBounds if empty
				suggestion := match.GetSuggestedReplacements()[0]
				suggestion = strings.ReplaceAll(suggestion, "{suggestion}", replacement)
				suggestion = strings.ReplaceAll(suggestion, "{Suggestion}", tools.UppercaseFirstChar(replacement))
				suggestion = strings.ReplaceAll(suggestion, "{SUGGESTION}", strings.ToUpper(replacement))
				ruleMatch.SetSuggestedReplacement(suggestion)
				return ruleMatch
			}
		}
	}
	return nil
}
