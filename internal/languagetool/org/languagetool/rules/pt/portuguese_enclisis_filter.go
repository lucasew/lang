package pt

import (
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// PortugueseEnclisisFilter ports org.languagetool.rules.pt.PortugueseEnclisisFilter.
// Verb form synthesis is optional via SynthesizeEnclisis (Java: PortugueseSynthesizer).
type PortugueseEnclisisFilter struct {
	// SynthesizeEnclisis(verbToken, verbPOS, pronounTag) → enclitic forms.
	// Nil → no synthesized forms (fail-closed empty suggestions, like empty synth output).
	SynthesizeEnclisis func(verbToken, verbPOS, pronounTag string) []string
}

func NewPortugueseEnclisisFilter() *PortugueseEnclisisFilter {
	return &PortugueseEnclisisFilter{}
}

// AcceptRuleMatch ports PortugueseEnclisisFilter.acceptRuleMatch.
// Args: verbPos, pronounPos (0-based pattern token indexes), convertToAccusative.
func (f *PortugueseEnclisisFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	verbPos, err1 := strconv.Atoi(arguments["verbPos"])
	pronounPos, err2 := strconv.Atoi(arguments["pronounPos"])
	if err1 != nil || err2 != nil {
		return nil
	}
	convertToAccusative := strings.EqualFold(arguments["convertToAccusative"], "true")
	if verbPos < 0 || verbPos >= len(patternTokens) || pronounPos < 0 || pronounPos >= len(patternTokens) {
		return nil
	}
	verbATR := patternTokens[verbPos]
	pronounATR := patternTokens[pronounPos]
	if verbATR == nil || pronounATR == nil {
		return nil
	}
	var prReadings []PronounTagReading
	for _, at := range pronounATR.GetReadings() {
		if at == nil {
			continue
		}
		pos := ""
		if at.GetPOSTag() != nil {
			pos = *at.GetPOSTag()
		}
		prReadings = append(prReadings, PronounTagReading{Token: at.GetToken(), POS: pos})
	}
	// Java: if readings empty, still use surface token for "nos" check via getToken on each reading.
	if len(prReadings) == 0 {
		prReadings = []PronounTagReading{{Token: pronounATR.GetToken()}}
	}
	pronounTags := f.PronounTags(prReadings, verbATR.GetToken(), convertToAccusative)
	if len(pronounTags) == 0 {
		return nil
	}
	// First verb reading with V* POS (Java break after first V).
	var suggestions []string
	for _, at := range verbATR.GetReadings() {
		if at == nil || at.GetPOSTag() == nil {
			continue
		}
		pos := *at.GetPOSTag()
		if !strings.HasPrefix(pos, "V") {
			continue
		}
		// Java synthesize uses AnalyzedToken token; use ATR surface for casing.
		suggestions = f.Suggest(VerbReading{Token: verbATR.GetToken(), POS: pos}, pronounTags)
		break
	}
	match.SetSuggestedReplacements(suggestions)
	return match
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
