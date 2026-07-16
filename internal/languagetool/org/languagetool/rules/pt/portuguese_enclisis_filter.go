package pt

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// PortugueseEnclisisFilter ports pronoun-tag extraction from
// org.languagetool.rules.pt.PortugueseEnclisisFilter.
// Verb form synthesis is optional via SynthesizeEnclisis.
type PortugueseEnclisisFilter struct {
	// SynthesizeEnclisis(verbToken, verbPOS, pronounTag) → enclitic forms.
	SynthesizeEnclisis func(verbToken, verbPOS, pronounTag string) []string
}

func NewPortugueseEnclisisFilter() *PortugueseEnclisisFilter {
	return &PortugueseEnclisisFilter{}
}

// PronounTagReading is one POS reading of a pronoun token.
type PronounTagReading struct {
	Token string
	POS   string
}

// PronounTags extracts PP tags from pronoun readings (with "nos" special case).
func (f *PortugueseEnclisisFilter) PronounTags(readings []PronounTagReading, verbText string, convertToAccusative bool) []string {
	var tags []string
	for _, pr := range readings {
		if pr.Token == "nos" {
			tags = append(tags, "PP1CPO00")
			if strings.HasSuffix(verbText, "m") || strings.HasSuffix(verbText, "ão") || strings.HasSuffix(verbText, "õe") {
				tags = append(tags, "PP3MPA00")
			}
			break
		}
		if pr.POS != "" && strings.HasPrefix(pr.POS, "PP") {
			pos := pr.POS
			if convertToAccusative {
				pos = convertPronounToAccusative(pos)
			}
			tags = append(tags, pos)
		}
	}
	return tags
}

func convertPronounToAccusative(pronounTag string) string {
	if strings.HasSuffix(pronounTag, "N00") {
		return pronounTag[:len(pronounTag)-3] + "A00"
	}
	return pronounTag
}

// VerbReading is one verb stem reading.
type VerbReading struct {
	Token string
	POS   string
}

// Suggest builds enclitic suggestions when SynthesizeEnclisis is set.
func (f *PortugueseEnclisisFilter) Suggest(verb VerbReading, pronounTags []string) []string {
	if f.SynthesizeEnclisis == nil || len(pronounTags) == 0 {
		return nil
	}
	if !strings.HasPrefix(verb.POS, "V") {
		return nil
	}
	isTitle := tools.IsCapitalizedWord(verb.Token)
	isAllCaps := tools.IsAllUppercase(verb.Token)
	seen := map[string]struct{}{}
	var out []string
	for _, ptag := range pronounTags {
		for _, form := range f.SynthesizeEnclisis(verb.Token, verb.POS, ptag) {
			if isTitle {
				form = tools.UppercaseFirstChar(form)
			} else if isAllCaps {
				form = strings.ToUpper(form)
			}
			if _, ok := seen[form]; ok {
				continue
			}
			seen[form] = struct{}{}
			out = append(out, form)
		}
	}
	return out
}
