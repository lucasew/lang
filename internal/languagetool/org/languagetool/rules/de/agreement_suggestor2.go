package de

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AgreementSuggestor2 ports org.languagetool.rules.de.AgreementSuggestor2.
// Create suggestions for German noun phrases that lack agreement.
type AgreementSuggestor2 struct {
	synthesizer      synthesis.Synthesizer
	determinerToken  *languagetool.AnalyzedTokenReadings
	adjToken1        *languagetool.AnalyzedTokenReadings
	adjToken2        *languagetool.AnalyzedTokenReadings
	nounToken        *languagetool.AnalyzedTokenReadings
	replacementType  ReplacementType
	prepositionToken *languagetool.AnalyzedTokenReadings
	skippedStr       string
	origPhrase       string
}

// Java templates (AgreementSuggestor2).
const (
	detTemplate = "ART:IND/DEF:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU"
	adjTemplate = "ADJ:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU:GRU:IND/DEF"
	pa1Template = "PA1:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU:GRU:IND/DEF:VER"
	pa2Template = "PA2:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU:GRU:IND/DEF:VER"
)

var (
	proPosTemplate = []string{
		"PRO:POS:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU:BEG",
		"PRO:POS:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU:B/S",
	}
	proDemTemplates = []string{
		"PRO:DEM:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU:BEG",
		"PRO:DEM:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU:B/S",
	}
	proIndTemplates = []string{
		"PRO:IND:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU:BEG",
		"PRO:IND:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU:B/S",
	}
	nounTemplates = []string{
		"SUB:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU",
		"SUB:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU:INF", // INF is for cases like "das Züchten" etc.
	}
	suggestorNumber = []string{"SIN", "PLU"}
	suggestorGender = []string{"MAS", "FEM", "NEU", "NOG"}
	suggestorCases  = []string{"NOM", "AKK", "DAT", "GEN"}
	// nounCases is the default when no preposition restricts cases.
	suggestorNounCases = []string{"NOM", "AKK", "DAT", "GEN"}
)

// skipSuggestions ports AgreementSuggestor2.skipSuggestions.
var skipSuggestions = map[string]struct{}{
	"unsren": {}, "unsrem": {}, "unsres": {}, "unsre": {}, "unsern": {}, "unserm": {}, "unsrer": {},
}

var (
	reDerselben  = regexp.MustCompile(`^([Dd]as|[Dd]er|[Dd]ie|[Dd]em|[Dd]es)selben?$`)
	reWelche     = regexp.MustCompile(`^[Ww]elche[nmsr]?$`)
	reHyphenTail = regexp.MustCompile(`.*-`) // Java String.replaceFirst(".*-", "")
)

// NewAgreementSuggestor2 ports AgreementSuggestor2(synthesizer, determinerToken, nounToken, replacementType)
// with null replacementType. Prefer NewAgreementSuggestor2Full for full Java constructors.
func NewAgreementSuggestor2(synth synthesis.Synthesizer, det, noun *languagetool.AnalyzedTokenReadings) *AgreementSuggestor2 {
	return NewAgreementSuggestor2Full(synth, det, nil, nil, noun, ReplNone)
}

// NewAgreementSuggestor2Adj ports AgreementSuggestor2(synth, det, adj1, noun, replacementType) since 5.4.
func NewAgreementSuggestor2Adj(synth synthesis.Synthesizer, det, adj1, noun *languagetool.AnalyzedTokenReadings, repl ReplacementType) *AgreementSuggestor2 {
	return NewAgreementSuggestor2Full(synth, det, adj1, nil, noun, repl)
}

// NewAgreementSuggestor2Full ports AgreementSuggestor2(synth, det, adj1, adj2, noun, replacementType) since 5.6.
func NewAgreementSuggestor2Full(synth synthesis.Synthesizer, det, adj1, adj2, noun *languagetool.AnalyzedTokenReadings, repl ReplacementType) *AgreementSuggestor2 {
	s := &AgreementSuggestor2{
		synthesizer:     synth,
		determinerToken: det,
		adjToken1:       adj1,
		adjToken2:       adj2,
		nounToken:       noun,
		replacementType: repl,
	}
	s.origPhrase = s.getOrigPhrase()
	return s
}

// WithAdjectives sets optional adjective tokens (test/wire convenience; recalculates origPhrase).
func (s *AgreementSuggestor2) WithAdjectives(adj1, adj2 *languagetool.AnalyzedTokenReadings) *AgreementSuggestor2 {
	if s != nil {
		s.adjToken1, s.adjToken2 = adj1, adj2
		s.origPhrase = s.getOrigPhrase()
	}
	return s
}

// WithReplacementType sets ins/zur contraction rewrite for suggestions.
func (s *AgreementSuggestor2) WithReplacementType(t ReplacementType) *AgreementSuggestor2 {
	if s != nil {
		s.replacementType = t
	}
	return s
}

// WithPreposition ports AgreementSuggestor2.setPreposition.
func (s *AgreementSuggestor2) WithPreposition(prep *languagetool.AnalyzedTokenReadings) *AgreementSuggestor2 {
	if s != nil {
		s.prepositionToken = prep
	}
	return s
}

// SetPreposition ports AgreementSuggestor2.setPreposition.
func (s *AgreementSuggestor2) SetPreposition(prep *languagetool.AnalyzedTokenReadings) {
	if s != nil {
		s.prepositionToken = prep
	}
}

// WithSkipped ports AgreementSuggestor2.setSkipped.
func (s *AgreementSuggestor2) WithSkipped(skipped string) *AgreementSuggestor2 {
	if s != nil {
		s.skippedStr = skipped
	}
	return s
}

// SetSkipped ports AgreementSuggestor2.setSkipped.
func (s *AgreementSuggestor2) SetSkipped(skipped string) {
	if s != nil {
		s.skippedStr = skipped
	}
}

// getOrigPhrase ports AgreementSuggestor2.getOrigPhrase.
func (s *AgreementSuggestor2) getOrigPhrase() string {
	if s == nil || s.determinerToken == nil || s.nounToken == nil {
		return ""
	}
	origElem := s.determinerToken.GetToken()
	if s.adjToken1 != nil {
		origElem += " " + s.adjToken1.GetToken()
	}
	if s.adjToken2 != nil {
		origElem += " " + s.adjToken2.GetToken()
	}
	origElem += " " + s.nounToken.GetToken()
	return origElem
}

// nounCasesForSuggestor ports AgreementSuggestor2.getNounCases.
// Case restriction comes only from prepositionToken (Java); Zur/Ins set that in getSuggestions.
func (s *AgreementSuggestor2) nounCasesForSuggestor() []string {
	if s != nil && s.prepositionToken != nil {
		// some prepositions require specific cases, so only generate those
		casesForToken := CasesForPreposition(s.prepositionToken.GetToken())
		if len(casesForToken) > 0 {
			var result []string
			for _, aCase := range casesForToken {
				// Java: aCase.name().toLowerCase() then toUpperCase for uniqueness
				val := strings.ToLower(grammaticalCaseName(aCase))
				upper := strings.ToUpper(val)
				if !containsStr(result, upper) {
					result = append(result, upper)
				}
			}
			return result
		}
	}
	return append([]string(nil), suggestorNounCases...)
}

func grammaticalCaseName(c GrammaticalCase) string {
	switch c {
	case CaseNom:
		return "NOM"
	case CaseAkk:
		return "AKK"
	case CaseDat:
		return "DAT"
	case CaseGen:
		return "GEN"
	default:
		return ""
	}
}

// agreementSuggestion ports AgreementSuggestor2.Suggestion.
type agreementSuggestion struct {
	phrase               string
	tokenLevelEdits      int
	charLevelCorrections int
}

// GetSuggestions ports getSuggestions() / getSuggestions(false).
func (s *AgreementSuggestor2) GetSuggestions() []string {
	return s.GetSuggestionsFiltered(false)
}

// GetSuggestionsFiltered ports getSuggestions(filter).
func (s *AgreementSuggestor2) GetSuggestionsFiltered(filter bool) []string {
	if s == nil || s.nounToken == nil || s.determinerToken == nil {
		return nil
	}
	// Java mutates prepositionToken for Zur/Ins (does not restore).
	if s.replacementType == ReplZur {
		pos := ""
		lem := "zu"
		s.prepositionToken = languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken("zu", &pos, &lem), 0)
	} else if s.replacementType == ReplIns {
		pos := ""
		lem := "zu" // Java: new AnalyzedToken("in", "", "zu")
		s.prepositionToken = languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken("in", &pos, &lem), 0)
	}
	sugs := s.getSuggestionsInternal()
	// sort so that suggestions with fewer edits come first (Java Collections.sort)
	sort.SliceStable(sugs, func(i, j int) bool {
		if sugs[i].tokenLevelEdits != sugs[j].tokenLevelEdits {
			return sugs[i].tokenLevelEdits < sugs[j].tokenLevelEdits
		}
		return sugs[i].charLevelCorrections < sugs[j].charLevelCorrections
	})
	sugs = s.addContraction(sugs)
	if filter {
		var filtered []agreementSuggestion
		prevCorrections := 0
		if len(sugs) > 0 {
			prevCorrections = sugs[0].tokenLevelEdits
		}
		hadRealSuggestions := false
		for _, suggestion := range sugs {
			if hadRealSuggestions && suggestion.tokenLevelEdits > prevCorrections {
				break
			}
			hadRealSuggestions = suggestion.tokenLevelEdits > 0
			filtered = append(filtered, suggestion)
		}
		out := make([]string, 0, len(filtered))
		for _, k := range filtered {
			out = append(out, k.phrase)
		}
		return out
	}
	out := make([]string, 0, len(sugs))
	for _, k := range sugs {
		out = append(out, k.phrase)
	}
	return out
}

func (s *AgreementSuggestor2) getSuggestionsInternal() []agreementSuggestion {
	nounCases := s.nounCasesForSuggestor()
	var result []agreementSuggestion
	for _, num := range suggestorNumber {
		for _, gen := range suggestorGender {
			for _, aCase := range suggestorCases {
				if !containsStr(nounCases, aCase) {
					continue
				}
				for _, detReading := range s.determinerToken.GetReadings() {
					if detReading == nil {
						continue
					}
					if gen == "NOG" {
						// needed to offer suggestions for e.g. "meinem Eltern" -> "meine Eltern, ..."
						detSynthesized := s.getDetOrPronounSynth(num, "MAS", aCase, detReading)
						adj1Synthesized := s.getAdjSynth(num, "MAS", aCase, s.adjToken1, detReading)
						adj2Synthesized := s.getAdjSynth(num, "MAS", aCase, s.adjToken2, detReading)
						nounSynthesized := s.getNounSynth(num, "NOG", aCase)
						s.combineSynth(&result, detSynthesized, adj1Synthesized, adj2Synthesized, nounSynthesized)
					} else {
						detSynthesized := s.getDetOrPronounSynth(num, gen, aCase, detReading)
						adj1Synthesized := s.getAdjSynth(num, gen, aCase, s.adjToken1, detReading)
						adj2Synthesized := s.getAdjSynth(num, gen, aCase, s.adjToken2, detReading)
						nounSynthesized := s.getNounSynth(num, gen, aCase)
						s.combineSynth(&result, detSynthesized, adj1Synthesized, adj2Synthesized, nounSynthesized)
					}
				}
			}
		}
	}
	return result
}

func containsStr(ss []string, v string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}

// getDetOrPronounSynth ports AgreementSuggestor2.getDetOrPronounSynth.
func (s *AgreementSuggestor2) getDetOrPronounSynth(num, gen, aCase string, detReading *languagetool.AnalyzedToken) []string {
	if detReading == nil || detReading.GetPOSTag() == nil {
		return []string{}
	}
	detPos := *detReading.GetPOSTag()
	isDef := strings.Contains(detPos, ":DEF:")
	tok := detReading.GetToken()
	var templates []string
	switch {
	case reDerselben.MatchString(tok):
		templates = []string{"PRO:DEM:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU"}
	case reWelche.MatchString(tok):
		templates = []string{"PRO:RIN:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU:B/S"}
	case strings.Contains(detPos, "ART:"):
		templates = []string{detTemplate}
	case strings.Contains(detPos, "PRO:POS:"):
		templates = append([]string(nil), proPosTemplate...)
	case strings.Contains(detPos, "PRO:DEM:"):
		templates = append([]string(nil), proDemTemplates...)
	case strings.Contains(detPos, "PRO:IND:"):
		templates = append([]string(nil), proIndTemplates...)
	case tok == "zur":
		templates = []string{detTemplate}
		pos := ""
		lem := "der"
		detReading = languagetool.NewAnalyzedToken("der", &pos, &lem)
		isDef = true
	case tok == "ins":
		templates = []string{detTemplate}
		pos := ""
		lem := "der"
		detReading = languagetool.NewAnalyzedToken("das", &pos, &lem)
		isDef = true
	default:
		return []string{}
	}
	var synthesized []string
	origFirst, _ := utf8.DecodeRuneInString(detReading.GetToken())
	for _, template := range templates {
		tpl := template
		if isDef {
			tpl = strings.Replace(tpl, "IND/DEF", "DEF", 1)
		} else {
			tpl = strings.Replace(tpl, "IND/DEF", "IND", 1)
		}
		pos := replaceSuggestorVars(tpl, num, gen, aCase)
		tmp := s.synthesizeToken(detReading, pos)
		for _, k := range tmp {
			// filter: starts with same first char (case-insensitive)
			if k == "" {
				continue
			}
			kr, _ := utf8.DecodeRuneInString(k)
			if strings.ToLower(string(kr)) != strings.ToLower(string(origFirst)) {
				continue
			}
			if _, skip := skipSuggestions[strings.ToLower(k)]; skip {
				continue
			}
			// don't suggest "dein" for "mein" etc. — keep original capitalisation
			if unicode.IsUpper(origFirst) {
				k = tools.UppercaseFirstChar(k)
			}
			synthesized = append(synthesized, k)
		}
	}
	return synthesized
}

// getAdjSynth ports AgreementSuggestor2.getAdjSynth.
func (s *AgreementSuggestor2) getAdjSynth(num, gen, aCase string, adjToken *languagetool.AnalyzedTokenReadings, detReading *languagetool.AnalyzedToken) []string {
	var adjSynthesized []string
	if adjToken != nil {
		for _, adjReading := range adjToken.GetReadings() {
			if adjReading == nil || adjReading.GetPOSTag() == nil || detReading == nil || detReading.GetPOSTag() == nil {
				continue
			}
			if adjReading.GetToken() == "meisten" && num == "SIN" {
				continue
			}
			adjPosTag := *adjReading.GetPOSTag()
			if strings.HasPrefix(adjPosTag, "ADV:") {
				// adverb, nothing to synthesize
				adjSynthesized = append(adjSynthesized, adjReading.GetToken())
				continue
			}
			detIsDef := strings.Contains(*detReading.GetPOSTag(), ":DEF:") || detReading.GetToken() == "ins"
			template := adjTemplate
			if strings.HasPrefix(adjPosTag, "PA1") {
				template = pa1Template
			} else if strings.HasPrefix(adjPosTag, "PA2") {
				template = pa2Template
			}
			if strings.Contains(adjPosTag, ":KOM:") {
				template = strings.Replace(template, ":GRU:", ":KOM:", 1)
			} else if strings.Contains(adjPosTag, ":SUP:") {
				template = strings.Replace(template, ":GRU:", ":SUP:", 1)
			}
			if detIsDef {
				template = strings.Replace(template, "IND/DEF", "DEF", 1)
			} else {
				template = strings.Replace(template, "IND/DEF", "IND", 1)
			}
			adjPos := replaceSuggestorVars(template, num, gen, aCase)
			for _, synthNoun := range s.synthesizeToken(adjReading, adjPos) {
				if !containsStr(adjSynthesized, synthNoun) {
					adjSynthesized = append(adjSynthesized, synthNoun)
				}
			}
		}
	} else {
		adjSynthesized = append(adjSynthesized, "") // noun phrase without an adjective
	}
	return adjSynthesized
}

// getNounSynth ports AgreementSuggestor2.getNounSynth.
func (s *AgreementSuggestor2) getNounSynth(num, gen, aCase string) []string {
	if s.nounToken == nil {
		return []string{}
	}
	var result []string
	for _, nounReading := range s.nounToken.GetReadings() {
		if nounReading == nil {
			continue
		}
		for _, nounTemplate := range nounTemplates {
			nounPos := replaceSuggestorVars(nounTemplate, num, gen, aCase)
			nounSynthesized := s.synthesizeToken(nounReading, nounPos)
			if len(nounSynthesized) == 0 && strings.Contains(nounReading.GetToken(), "-") {
				// hyphen compound: inflect last part only
				// Java: firstPart from nounReading; lastTokenPart from nounToken.getToken()
				tok := nounReading.GetToken()
				idx := strings.LastIndex(tok, "-")
				firstPart := tok[:idx+1]
				// nounToken.getToken().replaceFirst(".*-", "")
				lastTokenPart := reHyphenTail.ReplaceAllString(s.nounToken.GetToken(), "")
				var lastLemma *string
				if nounReading.GetLemma() != nil {
					// lemma.replaceFirst(".*-", "")
					l := reHyphenTail.ReplaceAllString(*nounReading.GetLemma(), "")
					lastLemma = &l
				}
				fakePos := "fake_value"
				fake := languagetool.NewAnalyzedToken(lastTokenPart, &fakePos, lastLemma)
				for _, lastPartInflected := range s.synthesizeToken(fake, nounPos) {
					result = append(result, firstPart+lastPartInflected)
				}
			} else {
				result = append(result, nounSynthesized...)
			}
		}
	}
	// remove forms like "Blutfluß" if there's also "Blutfluss" - must be old spelling then
	oldSpelling := map[string]struct{}{}
	for _, k := range result {
		if strings.Contains(k, "ss") {
			oldSpelling[strings.ReplaceAll(k, "ss", "ß")] = struct{}{}
		}
	}
	var filtered []string
	for _, k := range result {
		if _, bad := oldSpelling[k]; bad {
			continue
		}
		filtered = append(filtered, k)
	}
	return filtered
}

// combineSynth ports AgreementSuggestor2.combineSynth.
// Empty adj slices (nil or len 0) yield no combinations — only explicit [""] from null adjToken.
func (s *AgreementSuggestor2) combineSynth(result *[]agreementSuggestion, detSyn, adj1Syn, adj2Syn, nounSyn []string) {
	for _, detSynthesizedElem := range detSyn {
		for _, adj1SynthesizedElem := range adj1Syn {
			for _, adj2SynthesizedElem := range adj2Syn {
				for _, nounSynthesizedElem := range nounSyn {
					elem := detSynthesizedElem
					if s.skippedStr != "" {
						elem += " " + s.skippedStr
					}
					if adj1SynthesizedElem != "" {
						elem += " " + adj1SynthesizedElem
					}
					if adj2SynthesizedElem != "" {
						elem += " " + adj2SynthesizedElem
					}
					elem += " " + nounSynthesizedElem
					edits := 0
					if detSynthesizedElem != s.determinerToken.GetToken() {
						edits++
					}
					if s.adjToken1 == nil || adj1SynthesizedElem == s.adjToken1.GetToken() {
						// 0
					} else {
						edits++
					}
					if s.adjToken2 == nil || adj2SynthesizedElem == s.adjToken2.GetToken() {
						// 0
					} else {
						edits++
					}
					if nounSynthesizedElem != s.nounToken.GetToken() {
						edits++
					}
					if edits == 0 {
						continue
					}
					charLevelEdits := agreementLevenshtein(elem, s.origPhrase)
					suggestion := agreementSuggestion{
						phrase:               elem,
						tokenLevelEdits:      edits,
						charLevelCorrections: charLevelEdits,
					}
					if !containsSuggestion(*result, suggestion) {
						*result = append(*result, suggestion)
					}
				}
			}
		}
	}
}

// containsSuggestion ports List.contains via Suggestion.equals (phrase + tokenLevelEdits).
func containsSuggestion(list []agreementSuggestion, s agreementSuggestion) bool {
	for _, x := range list {
		if x.tokenLevelEdits == s.tokenLevelEdits && x.phrase == s.phrase {
			return true
		}
	}
	return false
}

// replaceSuggestorVars ports AgreementSuggestor2.replaceVars.
func replaceSuggestorVars(template, num, gen, aCase string) string {
	template = strings.Replace(template, "SIN/PLU", num, 1)
	template = strings.Replace(template, "MAS/FEM/NEU", gen, 1)
	template = strings.Replace(template, "NOM/AKK/DAT/GEN", aCase, 1)
	return template
}

func (s *AgreementSuggestor2) synthesizeToken(tok *languagetool.AnalyzedToken, pos string) []string {
	if s == nil || s.synthesizer == nil || tok == nil || pos == "" {
		return nil
	}
	forms, err := s.synthesizer.Synthesize(tok, pos)
	if err != nil {
		return nil
	}
	return forms
}

// addContraction ports AgreementSuggestor2.addContraction.
// Zur mutates phrases; Ins removes non-matching entries (Java Iterator.remove).
func (s *AgreementSuggestor2) addContraction(sugs []agreementSuggestion) []agreementSuggestion {
	if s == nil {
		return sugs
	}
	return applyReplacementContractionsOnSuggestions(sugs, s.replacementType)
}

// applyReplacementContractions ports addContraction for tests that only have phrase strings.
func applyReplacementContractions(phrases []string, t ReplacementType) []string {
	if t == ReplNone || len(phrases) == 0 {
		return phrases
	}
	sugs := make([]agreementSuggestion, len(phrases))
	for i, p := range phrases {
		sugs[i] = agreementSuggestion{phrase: p}
	}
	sugs = applyReplacementContractionsOnSuggestions(sugs, t)
	out := make([]string, 0, len(sugs))
	for _, s := range sugs {
		out = append(out, s.phrase)
	}
	return out
}

func applyReplacementContractionsOnSuggestions(sugs []agreementSuggestion, t ReplacementType) []agreementSuggestion {
	if t == ReplNone || len(sugs) == 0 {
		return sugs
	}
	var out []agreementSuggestion
	for _, sug := range sugs {
		p := sug.phrase
		switch t {
		case ReplZur:
			switch {
			case strings.HasPrefix(p, "der"):
				sug.phrase = strings.Replace(p, "der", "zur", 1)
			case strings.HasPrefix(p, "den"):
				sug.phrase = strings.Replace(p, "den", "zu", 1)
			case strings.HasPrefix(p, "dem"):
				sug.phrase = strings.Replace(p, "dem", "zum", 1)
			}
			out = append(out, sug)
		case ReplIns:
			switch {
			case strings.HasPrefix(p, "das"):
				sug.phrase = strings.Replace(p, "das", "ins", 1)
				out = append(out, sug)
			case strings.HasPrefix(p, "dem"):
				sug.phrase = strings.Replace(p, "dem", "im", 1)
				out = append(out, sug)
			case strings.HasPrefix(p, "den"):
				sug.phrase = strings.Replace(p, "den", "in den", 1)
				out = append(out, sug)
			case strings.HasPrefix(p, "die"):
				sug.phrase = strings.Replace(p, "die", "in die", 1)
				out = append(out, sug)
			default:
				// Java removes non-matching Ins suggestions
			}
		default:
			out = append(out, sug)
		}
	}
	return out
}

// agreementLevenshtein ports Apache Commons LevenshteinDistance.getDefaultInstance().apply
// (character-level, not rune-level: Java operates on UTF-16 code units via CharSequence).
// For German agreement phrases (BMP only) rune == char equivalence holds.
func agreementLevenshtein(a, b string) int {
	if a == b {
		return 0
	}
	// Java String length is UTF-16 code units; for BMP German text same as runes.
	ra := []rune(a)
	rb := []rune(b)
	la, lb := len(ra), len(rb)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	cur := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		cur[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := cur[j-1] + 1
			sub := prev[j-1] + cost
			m := del
			if ins < m {
				m = ins
			}
			if sub < m {
				m = sub
			}
			cur[j] = m
		}
		prev, cur = cur, prev
	}
	return prev[lb]
}

func strPtr(s string) *string { return &s }

// lemmaOf is retained for other callers in this package.
func lemmaOf(r *languagetool.AnalyzedTokenReadings) string {
	if r == nil {
		return ""
	}
	for _, t := range r.GetReadings() {
		if t != nil && t.GetLemma() != nil && *t.GetLemma() != "" {
			return *t.GetLemma()
		}
	}
	return r.GetToken()
}

func joinNonEmpty(parts ...string) string {
	var b []string
	for _, p := range parts {
		p = tools.JavaStringTrim(p)
		if p != "" {
			b = append(b, p)
		}
	}
	return strings.Join(b, " ")
}
