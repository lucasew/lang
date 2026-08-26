package pt

import (
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	ptsynthesis "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/pt"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// PortugueseEnclisisFilter ports org.languagetool.rules.pt.PortugueseEnclisisFilter.
//
// Java getSynthesizer() always returns PortugueseSynthesizer.INSTANCE. Production
// path uses INSTANCE by default; SetSynthesizer is an injectable override for tests.
//
// Edge mappings (documented Go twins of Java throw/NPE):
//   - invalid verbPos/pronounPos parse: Java NumberFormatException → return nil (fail-closed)
//   - empty pronounTags: Java return null
//   - synthesize error: Java IOException → skip forms for that call (fail-closed empty contrib)
//   - nil filter/match: Java NPE → return nil
//   - convertToAccusative: Java Boolean.parseBoolean (true only for "true", ignore case)
type PortugueseEnclisisFilter struct {
	// synth overrides getSynthesizer. Nil → PortugueseSynthesizer.INSTANCE (Java default).
	synth synthesis.Synthesizer
}

// NewPortugueseEnclisisFilter ports the Java no-arg constructor.
func NewPortugueseEnclisisFilter() *PortugueseEnclisisFilter {
	return &PortugueseEnclisisFilter{}
}

// getSynthesizer ports PortugueseEnclisisFilter.getSynthesizer.
// Java: return PortugueseSynthesizer.INSTANCE;
func (f *PortugueseEnclisisFilter) getSynthesizer() synthesis.Synthesizer {
	if f != nil && f.synth != nil {
		return f.synth
	}
	return ptsynthesis.INSTANCE
}

// SetSynthesizer overrides the synthesizer (tests / host injection only).
func (f *PortugueseEnclisisFilter) SetSynthesizer(s synthesis.Synthesizer) {
	if f != nil {
		f.synth = s
	}
}

// convertPronounToAccusative ports PortugueseEnclisisFilter.convertPronounToAccusative.
// Used when we have a pronoun like "eles", that is frequently used as a direct object,
// but we need it to be "os", which is tagged as the accusative form of "eles".
func convertPronounToAccusative(pronounTag string) string {
	if strings.HasSuffix(pronounTag, "N00") {
		return pronounTag[:len(pronounTag)-3] + "A00"
	}
	return pronounTag
}

// getPronounTags ports PortugueseEnclisisFilter.getPronounTags.
// Extract pronoun tags from pronoun tokens. We cannot just get all tags uncritically
// because of the 'nos' issue.
func getPronounTags(pronounReadings *languagetool.AnalyzedTokenReadings, verbText string, convertToAccusative bool) []string {
	pronounTags := make([]string, 0)
	if pronounReadings == nil {
		return pronounTags
	}
	for _, pronounToken := range pronounReadings.GetReadings() {
		if pronounToken == nil {
			continue
		}
		pronounText := pronounToken.GetToken()
		if pronounText == "nos" {
			pronounTags = append(pronounTags, "PP1CPO00")
			if strings.HasSuffix(verbText, "m") || strings.HasSuffix(verbText, "ão") || strings.HasSuffix(verbText, "õe") {
				pronounTags = append(pronounTags, "PP3MPA00")
			}
			break
		}
		posTagPtr := pronounToken.GetPOSTag()
		if posTagPtr != nil && strings.HasPrefix(*posTagPtr, "PP") {
			posTag := *posTagPtr
			if convertToAccusative {
				posTag = convertPronounToAccusative(posTag)
			}
			pronounTags = append(pronounTags, posTag)
		}
	}
	return pronounTags
}

// getVerbForms ports PortugueseEnclisisFilter.getVerbForms.
// Java HashSet then ArrayList — order is not stable; tests must use set membership.
func (f *PortugueseEnclisisFilter) getVerbForms(verbStemTokenReadings *languagetool.AnalyzedTokenReadings, pronounTags []string) []string {
	suggestions := map[string]struct{}{}
	if verbStemTokenReadings == nil {
		return []string{}
	}
	isTitleCase := tools.IsCapitalizedWord(verbStemTokenReadings.GetToken())
	isAllCaps := tools.IsAllUppercase(verbStemTokenReadings.GetToken())
	for _, at := range verbStemTokenReadings.GetReadings() {
		if at == nil {
			continue
		}
		posTagPtr := at.GetPOSTag()
		if posTagPtr == nil || !strings.HasPrefix(*posTagPtr, "V") {
			continue
		}
		posTag := *posTagPtr
		for _, pronounTag := range pronounTags {
			enclisisTag := posTag + ":" + pronounTag
			// Java: getSynthesizer().synthesize(at, enclisisTag) — full AnalyzedToken.
			forms, err := f.getSynthesizer().Synthesize(at, enclisisTag)
			if err != nil {
				// Java IOException propagates; Go RuleFilter has no error — fail-closed skip.
				continue
			}
			for _, form := range forms {
				if isTitleCase {
					form = tools.UppercaseFirstChar(form)
				} else if isAllCaps {
					form = strings.ToUpper(form)
				}
				suggestions[form] = struct{}{}
			}
		}
		break
	}
	out := make([]string, 0, len(suggestions))
	for s := range suggestions {
		out = append(out, s)
	}
	return out
}

// AcceptRuleMatch ports PortugueseEnclisisFilter.acceptRuleMatch.
// Args: verbPos, pronounPos (pattern token indexes), convertToAccusative.
func (f *PortugueseEnclisisFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	// Java NPE on null this/match; Go nil-receiver safety for host calls.
	if f == nil || match == nil {
		return nil
	}
	// Java: Integer.parseInt — NumberFormatException; Go fail-closed.
	verbPos, err1 := strconv.Atoi(arguments["verbPos"])
	pronounPos, err2 := strconv.Atoi(arguments["pronounPos"])
	if err1 != nil || err2 != nil {
		return nil
	}
	// Java: Boolean.parseBoolean — true only for "true" (ignore case).
	convertToAccusative := strings.EqualFold(arguments["convertToAccusative"], "true")
	// Java: patternTokens[verbPos] / patternTokens[pronounPos] — OOB throws.
	if verbPos < 0 || verbPos >= len(patternTokens) || pronounPos < 0 || pronounPos >= len(patternTokens) {
		// Documented fail-closed host safety (Java would AIOOB).
		return nil
	}
	verbStemTokenReadings := patternTokens[verbPos]
	pronounTokenReadings := patternTokens[pronounPos]
	pronounTags := getPronounTags(pronounTokenReadings, verbStemTokenReadings.GetToken(), convertToAccusative)
	if len(pronounTags) == 0 {
		return nil
	}
	match.SetSuggestedReplacements(f.getVerbForms(verbStemTokenReadings, pronounTags))
	return match
}
