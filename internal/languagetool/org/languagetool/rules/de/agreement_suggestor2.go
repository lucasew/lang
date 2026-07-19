package de

import (
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AgreementSuggestor2 ports rules.de.AgreementSuggestor2.
// Generates alternative determiner/adjective/noun forms via synthesizer for
// case/number/gender combinations, ranked by token-level then char-level edits.
type AgreementSuggestor2 struct {
	Synth      synthesis.Synthesizer
	Determiner *languagetool.AnalyzedTokenReadings
	Adj1       *languagetool.AnalyzedTokenReadings
	Adj2       *languagetool.AnalyzedTokenReadings
	Noun       *languagetool.AnalyzedTokenReadings
	// ReplType ports AgreementRule.ReplacementType for ins/zur contraction rewrite.
	ReplType ReplacementType
	// Preposition ports setPreposition — restricts cases via PrepositionToCases.
	Preposition *languagetool.AnalyzedTokenReadings
	// SkippedStr ports setSkipped (modifiers between det and adj/noun).
	SkippedStr string
	// SkipSuggestions filters generated forms.
	SkipSuggestions map[string]struct{}
	origPhrase      string
}

// Java templates (AgreementSuggestor2).
const (
	detTemplate  = "ART:IND/DEF:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU"
	adjTemplate  = "ADJ:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU:GRU:IND/DEF"
	pa1Template  = "PA1:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU:GRU:IND/DEF:VER"
	pa2Template  = "PA2:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU:GRU:IND/DEF:VER"
)

var (
	proPosTemplates = []string{
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
		"SUB:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU:INF",
	}
	suggestorNumbers = []string{"SIN", "PLU"}
	suggestorGenders = []string{"MAS", "FEM", "NEU", "NOG"}
	suggestorCases   = []string{"NOM", "AKK", "DAT", "GEN"}
)

var defaultSkipSuggestions = map[string]struct{}{
	"unsren": {}, "unsrem": {}, "unsres": {}, "unsre": {}, "unsern": {}, "unserm": {}, "unsrer": {},
}

var (
	reDerselben = regexp.MustCompile(`^([Dd]as|[Dd]er|[Dd]ie|[Dd]em|[Dd]es)selben?$`)
	reWelche    = regexp.MustCompile(`^[Ww]elche[nmsr]?$`)
)

func NewAgreementSuggestor2(synth synthesis.Synthesizer, det, noun *languagetool.AnalyzedTokenReadings) *AgreementSuggestor2 {
	s := &AgreementSuggestor2{
		Synth:           synth,
		Determiner:      det,
		Noun:            noun,
		SkipSuggestions: defaultSkipSuggestions,
	}
	s.origPhrase = s.computeOrigPhrase()
	return s
}

// WithAdjectives sets optional adjective tokens.
func (s *AgreementSuggestor2) WithAdjectives(adj1, adj2 *languagetool.AnalyzedTokenReadings) *AgreementSuggestor2 {
	if s != nil {
		s.Adj1, s.Adj2 = adj1, adj2
		s.origPhrase = s.computeOrigPhrase()
	}
	return s
}

// WithReplacementType sets ins/zur contraction rewrite for suggestions.
func (s *AgreementSuggestor2) WithReplacementType(t ReplacementType) *AgreementSuggestor2 {
	if s != nil {
		s.ReplType = t
	}
	return s
}

// WithPreposition ports AgreementSuggestor2.setPreposition.
func (s *AgreementSuggestor2) WithPreposition(prep *languagetool.AnalyzedTokenReadings) *AgreementSuggestor2 {
	if s != nil {
		s.Preposition = prep
	}
	return s
}

// WithSkipped ports AgreementSuggestor2.setSkipped.
func (s *AgreementSuggestor2) WithSkipped(skipped string) *AgreementSuggestor2 {
	if s != nil {
		s.SkippedStr = skipped
	}
	return s
}

func (s *AgreementSuggestor2) computeOrigPhrase() string {
	if s == nil || s.Determiner == nil || s.Noun == nil {
		return ""
	}
	parts := []string{s.Determiner.GetToken()}
	if s.Adj1 != nil {
		parts = append(parts, s.Adj1.GetToken())
	}
	if s.Adj2 != nil {
		parts = append(parts, s.Adj2.GetToken())
	}
	parts = append(parts, s.Noun.GetToken())
	return strings.Join(parts, " ")
}

// nounCasesForSuggestor ports AgreementSuggestor2.getNounCases.
func (s *AgreementSuggestor2) nounCasesForSuggestor() []string {
	all := append([]string(nil), suggestorCases...)
	if s == nil {
		return all
	}
	// Java ReplacementType temporarily sets prep; Ins uses "in" (not in Java PrepositionToCases → all).
	// Zur sets "zu" → DAT. We apply ReplType before consulting prep map.
	if s.ReplType == ReplIns {
		return all
	}
	if s.ReplType == ReplZur {
		return []string{"DAT"}
	}
	if s.Preposition == nil {
		return all
	}
	cases := CasesForPreposition(s.Preposition.GetToken())
	if len(cases) == 0 {
		return all
	}
	var out []string
	seen := map[string]struct{}{}
	for _, c := range cases {
		name := grammaticalCaseName(c)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	if len(out) == 0 {
		return all
	}
	return out
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
	phrase            string
	tokenLevelEdits   int
	charLevelEdits    int
}

// GetSuggestions ports getSuggestions(false).
func (s *AgreementSuggestor2) GetSuggestions() []string {
	return s.GetSuggestionsFiltered(false)
}

// GetSuggestionsFiltered ports getSuggestions(filter).
func (s *AgreementSuggestor2) GetSuggestionsFiltered(filter bool) []string {
	if s == nil || s.Noun == nil || s.Determiner == nil {
		return nil
	}
	// Java: temporary prep rewrite for Zur/Ins before internal generation
	savedPrep := s.Preposition
	if s.ReplType == ReplZur {
		lem := "zu"
		s.Preposition = languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken("zu", nil, &lem), 0)
	} else if s.ReplType == ReplIns {
		lem := "zu" // Java lemma is "zu" for the synthetic "in" token
		s.Preposition = languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken("in", nil, &lem), 0)
	}
	sugs := s.getSuggestionsInternal()
	s.Preposition = savedPrep

	sort.SliceStable(sugs, func(i, j int) bool {
		if sugs[i].tokenLevelEdits != sugs[j].tokenLevelEdits {
			return sugs[i].tokenLevelEdits < sugs[j].tokenLevelEdits
		}
		return sugs[i].charLevelEdits < sugs[j].charLevelEdits
	})
	// contractions mutate phrase surfaces
	sugs = applyReplacementContractionsOnSuggestions(sugs, s.ReplType)
	if filter {
		var filtered []agreementSuggestion
		prev := 0
		if len(sugs) > 0 {
			prev = sugs[0].tokenLevelEdits
		}
		hadReal := false
		for _, sug := range sugs {
			if hadReal && sug.tokenLevelEdits > prev {
				break
			}
			hadReal = sug.tokenLevelEdits > 0
			filtered = append(filtered, sug)
		}
		sugs = filtered
	}
	out := make([]string, 0, len(sugs))
	for _, sug := range sugs {
		out = append(out, sug.phrase)
	}
	return out
}

func (s *AgreementSuggestor2) getSuggestionsInternal() []agreementSuggestion {
	nounCases := s.nounCasesForSuggestor()
	// Re-apply ReplType case restriction after prep was set for Ins/Zur in GetSuggestionsFiltered:
	// nounCasesForSuggestor already handles ReplType first.
	var result []agreementSuggestion
	seen := map[string]struct{}{} // phrase+edits key like Java equals
	for _, num := range suggestorNumbers {
		for _, gen := range suggestorGenders {
			for _, aCase := range suggestorCases {
				if !containsStr(nounCases, aCase) {
					continue
				}
				for _, detReading := range s.Determiner.GetReadings() {
					if detReading == nil {
						continue
					}
					genForDet := gen
					if gen == "NOG" {
						// Java: det/adj as MAS, noun as NOG
						genForDet = "MAS"
					}
					detSyn := s.getDetOrPronounSynth(num, genForDet, aCase, detReading)
					adj1Syn := s.getAdjSynth(num, genForDet, aCase, s.Adj1, detReading)
					adj2Syn := s.getAdjSynth(num, genForDet, aCase, s.Adj2, detReading)
					nounGen := gen
					if gen == "NOG" {
						nounGen = "NOG"
					}
					nounSyn := s.getNounSynth(num, nounGen, aCase)
					s.combineSynth(&result, seen, detSyn, adj1Syn, adj2Syn, nounSyn)
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

func (s *AgreementSuggestor2) getDetOrPronounSynth(num, gen, aCase string, detReading *languagetool.AnalyzedToken) []string {
	if detReading == nil || detReading.GetPOSTag() == nil {
		return nil
	}
	detPos := *detReading.GetPOSTag()
	isDef := strings.Contains(detPos, ":DEF:")
	tok := detReading.GetToken()
	var templates []string
	synthReading := detReading
	switch {
	case reDerselben.MatchString(tok):
		templates = []string{"PRO:DEM:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU"}
	case reWelche.MatchString(tok):
		templates = []string{"PRO:RIN:NOM/AKK/DAT/GEN:SIN/PLU:MAS/FEM/NEU:B/S"}
	case strings.Contains(detPos, "ART:"):
		templates = []string{detTemplate}
	case strings.Contains(detPos, "PRO:POS:"):
		templates = append([]string(nil), proPosTemplates...)
	case strings.Contains(detPos, "PRO:DEM:"):
		templates = append([]string(nil), proDemTemplates...)
	case strings.Contains(detPos, "PRO:IND:"):
		templates = append([]string(nil), proIndTemplates...)
	case tok == "zur":
		templates = []string{detTemplate}
		lem := "der"
		synthReading = languagetool.NewAnalyzedToken("der", nil, &lem)
		isDef = true
	case tok == "ins":
		templates = []string{detTemplate}
		lem := "der"
		synthReading = languagetool.NewAnalyzedToken("das", nil, &lem)
		isDef = true
	default:
		return nil
	}
	var synthesized []string
	origFirst, _ := utf8.DecodeRuneInString(synthReading.GetToken())
	for _, template := range templates {
		tpl := template
		if isDef {
			tpl = strings.Replace(tpl, "IND/DEF", "DEF", 1)
		} else {
			tpl = strings.Replace(tpl, "IND/DEF", "IND", 1)
		}
		pos := replaceSuggestorVars(tpl, num, gen, aCase)
		tmp := s.synthesizeToken(synthReading, pos)
		for _, k := range tmp {
			if k == "" {
				continue
			}
			if s.skip(k) {
				continue
			}
			// first char must match (don't suggest "dein" for "mein")
			kr, _ := utf8.DecodeRuneInString(k)
			if strings.ToLower(string(kr)) != strings.ToLower(string(origFirst)) {
				continue
			}
			if unicodeIsUpper(origFirst) {
				k = tools.UppercaseFirstChar(k)
			}
			synthesized = append(synthesized, k)
		}
	}
	return synthesized
}

func unicodeIsUpper(r rune) bool {
	return r >= 'A' && r <= 'Z' || r == 'Ä' || r == 'Ö' || r == 'Ü'
}

func (s *AgreementSuggestor2) getAdjSynth(num, gen, aCase string, adjToken *languagetool.AnalyzedTokenReadings, detReading *languagetool.AnalyzedToken) []string {
	if adjToken == nil {
		return []string{""} // noun phrase without adjective
	}
	if detReading == nil || detReading.GetPOSTag() == nil {
		return nil
	}
	detPos := *detReading.GetPOSTag()
	detIsDef := strings.Contains(detPos, ":DEF:") || detReading.GetToken() == "ins"
	var adjSynthesized []string
	seen := map[string]struct{}{}
	for _, adjReading := range adjToken.GetReadings() {
		if adjReading == nil || adjReading.GetPOSTag() == nil {
			continue
		}
		adjPosTag := *adjReading.GetPOSTag()
		if adjReading.GetToken() == "meisten" && num == "SIN" {
			continue
		}
		if strings.HasPrefix(adjPosTag, "ADV:") {
			tok := adjReading.GetToken()
			if _, ok := seen[tok]; !ok {
				seen[tok] = struct{}{}
				adjSynthesized = append(adjSynthesized, tok)
			}
			continue
		}
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
		for _, form := range s.synthesizeToken(adjReading, adjPos) {
			if form == "" {
				continue
			}
			if _, ok := seen[form]; ok {
				continue
			}
			seen[form] = struct{}{}
			adjSynthesized = append(adjSynthesized, form)
		}
	}
	return adjSynthesized
}

func (s *AgreementSuggestor2) getNounSynth(num, gen, aCase string) []string {
	if s.Noun == nil {
		return nil
	}
	var result []string
	for _, nounReading := range s.Noun.GetReadings() {
		if nounReading == nil {
			continue
		}
		for _, nounTemplate := range nounTemplates {
			nounPos := replaceSuggestorVars(nounTemplate, num, gen, aCase)
			nounSynthesized := s.synthesizeToken(nounReading, nounPos)
			if len(nounSynthesized) == 0 && strings.Contains(nounReading.GetToken(), "-") {
				// hyphen compound: inflect last part only
				tok := nounReading.GetToken()
				idx := strings.LastIndex(tok, "-")
				firstPart := tok[:idx+1]
				lastTokenPart := tok[idx+1:]
				var lastLemma *string
				if nounReading.GetLemma() != nil {
					lem := *nounReading.GetLemma()
					if i := strings.LastIndex(lem, "-"); i >= 0 {
						l := lem[i+1:]
						lastLemma = &l
					} else {
						lastLemma = &lem
					}
				}
				fake := languagetool.NewAnalyzedToken(lastTokenPart, strPtr("fake_value"), lastLemma)
				for _, lastPart := range s.synthesizeToken(fake, nounPos) {
					result = append(result, firstPart+lastPart)
				}
			} else {
				result = append(result, nounSynthesized...)
			}
		}
	}
	// remove ß-forms when ss form also present (old spelling)
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

func strPtr(s string) *string { return &s }

func (s *AgreementSuggestor2) combineSynth(result *[]agreementSuggestion, seen map[string]struct{},
	detSyn, adj1Syn, adj2Syn, nounSyn []string) {
	if len(detSyn) == 0 || len(nounSyn) == 0 {
		return
	}
	// empty adj arrays mean "no combination" in Java (except explicit "")
	if adj1Syn == nil {
		adj1Syn = []string{""}
	}
	if adj2Syn == nil {
		adj2Syn = []string{""}
	}
	detTok := s.Determiner.GetToken()
	var adj1Tok, adj2Tok string
	if s.Adj1 != nil {
		adj1Tok = s.Adj1.GetToken()
	}
	if s.Adj2 != nil {
		adj2Tok = s.Adj2.GetToken()
	}
	nounTok := s.Noun.GetToken()
	for _, det := range detSyn {
		for _, adj1 := range adj1Syn {
			for _, adj2 := range adj2Syn {
				for _, noun := range nounSyn {
					elem := det
					if s.SkippedStr != "" {
						elem += " " + s.SkippedStr
					}
					if adj1 != "" {
						elem += " " + adj1
					}
					if adj2 != "" {
						elem += " " + adj2
					}
					elem += " " + noun
					edits := 0
					if det != detTok {
						edits++
					}
					if s.Adj1 != nil && adj1 != adj1Tok {
						edits++
					}
					if s.Adj2 != nil && adj2 != adj2Tok {
						edits++
					}
					if noun != nounTok {
						edits++
					}
					if edits == 0 {
						continue
					}
					charEdits := agreementLevenshtein(elem, s.origPhrase)
					key := elem + "\x00" + agreementItoa(edits)
					if _, ok := seen[key]; ok {
						continue
					}
					seen[key] = struct{}{}
					*result = append(*result, agreementSuggestion{
						phrase:          elem,
						tokenLevelEdits: edits,
						charLevelEdits:  charEdits,
					})
				}
			}
		}
	}
}

func agreementItoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func replaceSuggestorVars(template, num, gen, aCase string) string {
	template = strings.Replace(template, "SIN/PLU", num, 1)
	template = strings.Replace(template, "MAS/FEM/NEU", gen, 1)
	template = strings.Replace(template, "NOM/AKK/DAT/GEN", aCase, 1)
	return template
}

func (s *AgreementSuggestor2) synthesizeToken(tok *languagetool.AnalyzedToken, pos string) []string {
	if s == nil || s.Synth == nil || tok == nil || pos == "" {
		return nil
	}
	forms, err := s.Synth.Synthesize(tok, pos)
	if err != nil {
		return nil
	}
	return forms
}

func (s *AgreementSuggestor2) skip(form string) bool {
	if form == "" {
		return false
	}
	_, ok := s.SkipSuggestions[strings.ToLower(form)]
	return ok
}

// applyReplacementContractions ports addContraction on phrase list (tests).
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
				sug.phrase = "zur" + strings.TrimPrefix(p, "der")
			case strings.HasPrefix(p, "den"):
				sug.phrase = "zu" + strings.TrimPrefix(p, "den")
			case strings.HasPrefix(p, "dem"):
				sug.phrase = "zum" + strings.TrimPrefix(p, "dem")
			}
			out = append(out, sug)
		case ReplIns:
			switch {
			case strings.HasPrefix(p, "das"):
				sug.phrase = "ins" + strings.TrimPrefix(p, "das")
				out = append(out, sug)
			case strings.HasPrefix(p, "dem"):
				sug.phrase = "im" + strings.TrimPrefix(p, "dem")
				out = append(out, sug)
			case strings.HasPrefix(p, "den"):
				sug.phrase = "in den" + strings.TrimPrefix(p, "den")
				out = append(out, sug)
			case strings.HasPrefix(p, "die"):
				sug.phrase = "in die" + strings.TrimPrefix(p, "die")
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

// agreementLevenshtein is char-level edit distance (Java LevenshteinDistance).
func agreementLevenshtein(a, b string) int {
	if a == b {
		return 0
	}
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
		p = strings.TrimSpace(p)
		if p != "" {
			b = append(b, p)
		}
	}
	return strings.Join(b, " ")
}
