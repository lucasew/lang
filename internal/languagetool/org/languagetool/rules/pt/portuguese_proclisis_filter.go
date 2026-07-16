package pt

import "strings"

// PortugueseProclisisFilter ports pronoun remapping from
// org.languagetool.rules.pt.PortugueseProclisisFilter.
// Verb form synthesis is provided via SynthesizeVerb (optional).
type PortugueseProclisisFilter struct {
	// SynthesizeVerb returns the non-enclitic verb form for a lemma/POS; nil uses old verb stem.
	SynthesizeVerb func(token, verbTag string) string
}

func NewPortugueseProclisisFilter() *PortugueseProclisisFilter {
	return &PortugueseProclisisFilter{}
}

// Suggest builds proclisis suggestions from an enclitic verb token like "dizer-lhe".
// readings are (token, posTag) pairs for the matched enclitic verb.
func (f *PortugueseProclisisFilter) Suggest(readings []struct{ Token, POS string }) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, at := range readings {
		posTag := at.POS
		if posTag == "" || !strings.HasPrefix(posTag, "V") || !strings.Contains(posTag, ":") {
			continue
		}
		oldToken := at.Token
		tagParts := strings.Split(posTag, ":")
		verbTag := tagParts[0]
		newVerb := oldToken
		if f.SynthesizeVerb != nil {
			if s := f.SynthesizeVerb(oldToken, verbTag); s != "" {
				newVerb = s
			}
		} else {
			// surface fallback: strip clitics after first hyphen
			if i := strings.Index(oldToken, "-"); i >= 0 {
				newVerb = oldToken[:i]
			}
		}
		parts := strings.Split(oldToken, "-")
		if len(parts) < 2 {
			continue
		}
		oldVerb, oldPronoun := parts[0], parts[1]
		// if synthesizer not used, prefer stripped stem
		if f.SynthesizeVerb == nil {
			newVerb = oldVerb
		}
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
		out := []string{"nos"}
		if strings.HasSuffix(oldVerb, "m") || strings.HasSuffix(oldVerb, "ão") || strings.HasSuffix(oldVerb, "õe") {
			out = append(out, "os")
		}
		return out
	default:
		return []string{oldPronoun}
	}
}
