package pt

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PortugueseProclisisFilter ports org.languagetool.rules.pt.PortugueseProclisisFilter.
// Verb form synthesis is provided via SynthesizeVerb (Java: PortugueseSynthesizer).
// Without SynthesizeVerb, that reading is skipped (fail closed — no hyphen-strip invent).
type PortugueseProclisisFilter struct {
	// SynthesizeVerb returns the non-enclitic verb form for a token/POS
	// (Java: getSynthesizer().synthesize(at, verbTag)[0]).
	SynthesizeVerb func(token, verbTag string) string
}

func NewPortugueseProclisisFilter() *PortugueseProclisisFilter {
	return &PortugueseProclisisFilter{}
}

// AcceptRuleMatch ports PortugueseProclisisFilter.acceptRuleMatch.
func (f *PortugueseProclisisFilter) AcceptRuleMatch(match *rules.RuleMatch, _ map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil || len(patternTokens) == 0 {
		return nil
	}
	enclitic := patternTokens[len(patternTokens)-1]
	if enclitic == nil {
		return nil
	}
	var readings []struct{ Token, POS string }
	for _, at := range enclitic.GetReadings() {
		if at == nil {
			continue
		}
		pos := ""
		if at.GetPOSTag() != nil {
			pos = *at.GetPOSTag()
		}
		readings = append(readings, struct{ Token, POS string }{Token: at.GetToken(), POS: pos})
	}
	match.SetSuggestedReplacements(f.Suggest(readings))
	return match
}

// Suggest builds proclisis suggestions from an enclitic verb token like "dizer-lhe".
func (f *PortugueseProclisisFilter) Suggest(readings []struct{ Token, POS string }) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, at := range readings {
		posTag := at.POS
		if posTag == "" || !strings.HasPrefix(posTag, "V") || !strings.Contains(posTag, ":") {
			continue
		}
		if f.SynthesizeVerb == nil {
			// Java always has PortugueseSynthesizer; fail closed without it.
			continue
		}
		oldToken := at.Token
		tagParts := strings.Split(posTag, ":")
		verbTag := tagParts[0]
		newVerb := f.SynthesizeVerb(oldToken, verbTag)
		if newVerb == "" {
			continue
		}
		parts := strings.Split(oldToken, "-")
		if len(parts) < 2 {
			continue
		}
		oldVerb, oldPronoun := parts[0], parts[1]
		for _, newPronoun := range proclisisPronounForms(oldPronoun, oldVerb) {
			s := newPronoun + " " + newVerb
			if _, ok := seen[s]; ok {
				continue
			}
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}

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
		if strings.HasSuffix(oldVerb, "m") || strings.HasSuffix(oldVerb, "ão") || strings.HasSuffix(oldVerb, "õe") {
			forms = append(forms, "os")
		}
		return forms
	default:
		return []string{oldPronoun}
	}
}
