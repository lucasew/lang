package pt

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	ptsynthesis "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/pt"
)

// PortugueseProclisisFilter ports org.languagetool.rules.pt.PortugueseProclisisFilter.
//
// Java getSynthesizer() always returns PortugueseSynthesizer.INSTANCE. Production
// path uses INSTANCE by default; SetSynthesizer is an injectable override for tests.
//
// Edge mappings (documented Go twins of Java throw/NPE):
//   - empty patternTokens: Java ArrayIndexOutOfBoundsException → panic
//   - synthesize empty / error: Java [0] AIOOB / IOException → skip that reading (fail-closed)
//   - token without hyphen: Java oldTokenParts[1] AIOOB → skip that reading (fail-closed)
//   - nil filter/match: Java NPE → return nil
// Happy path always match.setSuggestedReplacements(...) and returns match (never null).
type PortugueseProclisisFilter struct {
	// synth overrides getSynthesizer. Nil → PortugueseSynthesizer.INSTANCE (Java default).
	synth synthesis.Synthesizer
}

// NewPortugueseProclisisFilter ports the Java no-arg constructor.
func NewPortugueseProclisisFilter() *PortugueseProclisisFilter {
	return &PortugueseProclisisFilter{}
}

// getSynthesizer ports PortugueseProclisisFilter.getSynthesizer.
// Java: return PortugueseSynthesizer.INSTANCE;
func (f *PortugueseProclisisFilter) getSynthesizer() synthesis.Synthesizer {
	if f != nil && f.synth != nil {
		return f.synth
	}
	return ptsynthesis.INSTANCE
}

// SetSynthesizer overrides the synthesizer (tests / host injection only).
func (f *PortugueseProclisisFilter) SetSynthesizer(s synthesis.Synthesizer) {
	if f != nil {
		f.synth = s
	}
}

// AcceptRuleMatch ports PortugueseProclisisFilter.acceptRuleMatch.
// Java always setSuggestedReplacements and returns match (never null).
func (f *PortugueseProclisisFilter) AcceptRuleMatch(match *rules.RuleMatch, _ map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	// Java NPE on null this/match; Go nil-receiver safety for host calls.
	if f == nil || match == nil {
		return nil
	}
	// Java: patternTokens[patternTokens.length - 1]
	if len(patternTokens) == 0 {
		panic("PortugueseProclisisFilter: empty patternTokens (Java ArrayIndexOutOfBoundsException)")
	}
	encliticVerbTokenReadings := patternTokens[len(patternTokens)-1]
	// Java HashSet — iteration order not stable; tests must use set membership.
	suggestions := map[string]struct{}{}
	for _, at := range encliticVerbTokenReadings.GetReadings() {
		if at == nil {
			continue
		}
		posTagPtr := at.GetPOSTag()
		// Java: if (posTag == null || !posTag.startsWith("V") || !posTag.contains(":")) continue;
		if posTagPtr == nil {
			continue
		}
		posTag := *posTagPtr
		if !strings.HasPrefix(posTag, "V") || !strings.Contains(posTag, ":") {
			continue
		}
		oldToken := at.GetToken()
		tagParts := strings.Split(posTag, ":")
		verbTag := tagParts[0]
		// Java: getSynthesizer().synthesize(at, verbTag)[0]
		// Pass full AnalyzedToken (lemma/token/pos) like Java.
		forms, err := f.getSynthesizer().Synthesize(at, verbTag)
		if err != nil || len(forms) == 0 {
			// Java throws or AIOOB on empty; skip reading (fail-closed mapping of empty synth).
			continue
		}
		newVerb := forms[0]
		oldTokenParts := strings.Split(oldToken, "-")
		// Java: oldTokenParts[0]/[1] — Includes only the relevant part of the verb;
		// for mesoclitics, we ignore the ending.
		if len(oldTokenParts) < 2 {
			// Java AIOOB; skip malformed token (fail-closed).
			continue
		}
		oldVerb := oldTokenParts[0]
		oldPronoun := oldTokenParts[1]
		for _, newPronoun := range proclisisPronounForms(oldPronoun, oldVerb) {
			suggestions[newPronoun+" "+newVerb] = struct{}{}
		}
	}
	// Java: match.setSuggestedReplacements(new ArrayList<>(suggestions)); return match;
	out := make([]string, 0, len(suggestions))
	for s := range suggestions {
		out = append(out, s)
	}
	match.SetSuggestedReplacements(out)
	return match
}

// proclisisPronounForms ports the Java switch on oldPronoun (exact cases).
func proclisisPronounForms(oldPronoun, oldVerb string) []string {
	switch oldPronoun {
	case "lo", "no":
		return []string{"o"}
	case "la", "na":
		return []string{"a"}
	case "los":
		return []string{"os"}
	case "las", "nas":
		return []string{"as"}
	case "nos":
		forms := []string{"nos"}
		// Java: if (oldVerb.endsWith("m") || oldVerb.endsWith("ão") || oldVerb.endsWith("õe"))
		if strings.HasSuffix(oldVerb, "m") || strings.HasSuffix(oldVerb, "ão") || strings.HasSuffix(oldVerb, "õe") {
			forms = append(forms, "os")
		}
		return forms
	default:
		return []string{oldPronoun}
	}
}
