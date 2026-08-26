package uk

import (
	"regexp"
	"sort"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
)

const TokenAgreementNounVerbRuleID = "UK_NOUN_VERB_INFLECTION_AGREEMENT"

// TokenAgreementNounVerbRule ports TokenAgreementNounVerbRule.
type TokenAgreementNounVerbRule struct {
	*tokenAgreementMatch
}

func NewTokenAgreementNounVerbRule() *TokenAgreementNounVerbRule {
	return NewTokenAgreementNounVerbRuleWithMessages(nil)
}

// NewTokenAgreementNounVerbRuleWithMessages ports the Java ctor (ResourceBundle messages).
func NewTokenAgreementNounVerbRuleWithMessages(messages map[string]string) *TokenAgreementNounVerbRule {
	r := &TokenAgreementNounVerbRule{}
	r.tokenAgreementMatch = &tokenAgreementMatch{
		ruleID:      TokenAgreementNounVerbRuleID,
		description: "Узгодження іменника та дієслова за родом, числом та особою",
		shortMsg:    "Узгодження іменника з дієсловом",
	}
	initTokenAgreementMeta(r.tokenAgreementMatch, messages)
	return r
}

var (
	nounVerbNazRE     = regexp.MustCompile(`noun.*:v_naz.*`)
	nounVerbFPVerbRE  = regexp.MustCompile(`verb.*:[fp]\b.*`)
	nounVerbPartRE    = regexp.MustCompile(`^part`)
	nounVerbAdjNazRE  = regexp.MustCompile(`^adj:.:(v_naz|v_kly).*`)
	nounVerbAdjZnaRE  = regexp.MustCompile(`^adj:m:v_zna:rinanim`)
	nounVerbPredInsRE = regexp.MustCompile(`^noninfl:(predic|insert).*`)
	nounVerbSkipToks  = map[string]struct{}{
		"не": {}, "б": {}, "би": {}, "бодай": {},
	}
	nounVerbAdjBlock = map[string]struct{}{
		"кожен": {}, "інший": {}, "старий": {}, "черговий": {},
	}
)

// isPredictOrInsertPOS ports PosTagHelper.isPredictOrInsert.
func isPredictOrInsertPOS(pos string) bool {
	return nounVerbPredInsRE.MatchString(pos)
}

// nounVerbState ports TokenAgreementNounVerbRule.State.
type nounVerbState struct {
	nounPos                   int
	nounTokenReadings         []string // POS tags of selected noun/яка readings
	nounAnalyzedTokenReadings *languagetool.AnalyzedTokenReadings
}

// Match ports TokenAgreementNounVerbRule.match.
func (r *TokenAgreementNounVerbRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch
	var state *nounVerbState

	// Java: for (int i = 1; i < tokens.length; i++)
	start := 1
	if len(tokens) > 0 && tokens[0] != nil && !tokens[0].IsSentenceStart() && firstPOS(tokens[0]) != "SENT_START" {
		start = 0
	}
	for i := start; i < len(tokens); i++ {
		tokenReadings := tokens[i]
		if tokenReadings == nil {
			state = nil
			continue
		}
		posTag0 := firstPOS(tokenReadings)
		if posTag0 == "" {
			state = nil
			continue
		}

		if state == nil {
			// no need to start checking on last token
			if i == len(tokens)-1 {
				continue
			}
		}

		cleanToken := cleanTokenSurface(tokenReadings)

		if HasPosTagRE(tokenReadings, nounVerbNazRE) || strings.EqualFold(cleanToken, "яка") {
			st, ok := collectNounVerbSubjectState(tokens, i)
			if !ok {
				state = nil
				continue
			}
			state = st
			continue
		}

		if state == nil {
			continue
		}

		if _, ok := nounVerbSkipToks[tokenReadings.GetToken()]; ok {
			continue
		}
		if hasPosTagPartAll(tokenReadings, "adv") {
			continue
		}

		// see if we get a following verb
		var verbTokenReadings []string
		clear := false
		for _, rdg := range tokenReadings.GetReadings() {
			if rdg == nil {
				continue
			}
			verbPosTag := ""
			if rdg.GetPOSTag() != nil {
				verbPosTag = *rdg.GetPOSTag()
			}
			if verbPosTag == "" || verbPosTag == "SENT_END" || verbPosTag == "PARA_END" {
				continue
			}
			if strings.HasPrefix(verbPosTag, "<") {
				verbTokenReadings = nil
				clear = true
				break
			}
			if strings.HasPrefix(verbPosTag, "verb") {
				verbTokenReadings = append(verbTokenReadings, verbPosTag)
			} else if isPredictOrInsertPOS(verbPosTag) {
				// ignore
			} else {
				verbTokenReadings = nil
				clear = true
				break
			}
		}
		if clear || len(verbTokenReadings) == 0 {
			state = nil
			continue
		}

		masterInflections := GetNounInflections(state.nounTokenReadings)
		slaveInflections := GetVerbInflections(verbTokenReadings)

		if !verbInflectionsOverlapLists(masterInflections, slaveInflections) {
			if IsNounVerbException(tokens, state.nounPos, i) {
				// Java: state.nounTokenReadings.clear(); break;
				state = nil
				break
			}

			// Java: LemmaHelper.hasLemma(state.nounTokenReadings, "який") on selected readings.
			kind := "іменник"
			if hasSelectedYakyi(state) {
				kind = "займенник"
			}

			msg := "Не узгоджено " + kind + " з дієсловом: \"" +
				firstSelectedToken(state) + "\" (" + formatVerbPersonInflections(masterInflections, true) +
				") і \"" + tokenReadings.GetToken() + "\" (" +
				formatVerbPersonInflections(slaveInflections, false) + ")"
			m := rules.NewRuleMatch(r, sentence, state.nounAnalyzedTokenReadings.GetStartPos(), tokenReadings.GetEndPos(), msg)
			m.ShortMessage = r.shortMsg
			ruleMatches = append(ruleMatches, m)
		}

		state = nil
	}
	return ruleMatches
}

func hasSelectedYakyi(state *nounVerbState) bool {
	if state == nil || state.nounAnalyzedTokenReadings == nil {
		return false
	}
	// Java: hasLemma(state.nounTokenReadings, "який") on selected readings only.
	for _, pos := range state.nounTokenReadings {
		if strings.Contains(pos, ":f:v_naz") {
			// check lemma on token readings with this pos
			for _, rdg := range state.nounAnalyzedTokenReadings.GetReadings() {
				if rdg == nil || rdg.GetPOSTag() == nil || rdg.GetLemma() == nil {
					continue
				}
				if *rdg.GetPOSTag() == pos && *rdg.GetLemma() == "який" {
					return true
				}
			}
		}
	}
	return false
}

func firstSelectedToken(state *nounVerbState) string {
	if state == nil || state.nounAnalyzedTokenReadings == nil {
		return ""
	}
	// Java: state.nounTokenReadings.get(0).getToken()
	return state.nounAnalyzedTokenReadings.GetToken()
}

// formatVerbPersonInflections ports TokenAgreementNounVerbRule.formatInflections.
func formatVerbPersonInflections(infs []VerbInflection, noun bool) string {
	_ = noun
	if len(infs) == 0 {
		return ""
	}
	sorted := append([]VerbInflection(nil), infs...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return verbInflectionOrder(sorted[i]) < verbInflectionOrder(sorted[j])
	})
	seen := map[string]struct{}{}
	var list []string
	for _, inf := range sorted {
		str := ""
		if inf.Gender != "" {
			str = taguk.GenderName(inf.Gender)
		} else {
			if inf.Person != "" {
				str = taguk.PersonName(inf.Person)
			}
			if inf.Plural != "" {
				if str != "" {
					str += " "
				}
				str += taguk.GenderName(inf.Plural)
			}
		}
		if str == "" {
			continue
		}
		if _, ok := seen[str]; ok {
			continue
		}
		seen[str] = struct{}{}
		list = append(list, str)
	}
	return strings.Join(list, ", ")
}

func verbInflectionOrder(inf VerbInflection) int {
	if inf.Gender == "" {
		return 0
	}
	if o, ok := genOrder[inf.Gender]; ok {
		return o
	}
	return 99
}

// collectNounVerbSubjectState ports the Java subject-state builder.
// ok=false → clear state (Java state=null).
func collectNounVerbSubjectState(tokens []*languagetool.AnalyzedTokenReadings, i int) (*nounVerbState, bool) {
	if i < 0 || i >= len(tokens) || tokens[i] == nil {
		return nil, false
	}
	tok := tokens[i]
	clean := cleanTokenSurface(tok)
	state := &nounVerbState{nounPos: -1}

	for _, rdg := range tok.GetReadings() {
		if rdg == nil {
			continue
		}
		nounPosTag := ""
		if rdg.GetPOSTag() != nil {
			nounPosTag = *rdg.GetPOSTag()
		}
		// can happen for words with \u0301 or \u00AD
		if nounPosTag == "" {
			continue
		}
		lem := ""
		if rdg.GetLemma() != nil {
			lem = *rdg.GetLemma()
		}

		if lem == "який" && strings.Contains(nounPosTag, ":f:v_naz") {
			state.nounPos = i
			state.nounTokenReadings = append(state.nounTokenReadings, nounPosTag)
			state.nounAnalyzedTokenReadings = tok
			continue
		}
		if strings.EqualFold(clean, "хто") {
			// ignore: про те, хто була ця клята Пандора
			if tokenSearchVerbFPSkipPart(tokens, i+1) >= 0 {
				return nil, false
			}
			if i < len(tokens)-1 && hasPosTagPartAll(tokens[i+1], "adv") &&
				tokenSearchVerbFPSkipPart(tokens, i+2) >= 0 {
				return nil, false
			}
		}
		if strings.HasPrefix(nounPosTag, "noun") && strings.Contains(nounPosTag, "v_naz") {
			state.nounPos = i
			state.nounTokenReadings = append(state.nounTokenReadings, nounPosTag)
			state.nounAnalyzedTokenReadings = tok
			continue
		}
		if strings.HasPrefix(nounPosTag, "noun") && strings.Contains(nounPosTag, "v_kly") {
			continue // ignore
		}
		if isPredictOrInsertPOS(nounPosTag) {
			continue
		}
		// Java: adj:.:(v_naz|v_kly).* || (adj:m:v_zna:rinanim && !prep) && !blocked
		// due to &&/|| precedence: A || ((B && !prep) && !blocked)
		if nounVerbAdjNazRE.MatchString(nounPosTag) {
			// adj readings not used for master inflections (Java adjTokenReadings only)
			continue
		}
		if strings.HasPrefix(nounPosTag, "adj:m:v_zna:rinanim") &&
			!(i > 0 && tokens[i-1] != nil && HasPosTagStart(tokens[i-1], "prep")) {
			lower := strings.ToLower(tok.GetToken())
			if _, block := nounVerbAdjBlock[lower]; block {
				return nil, false
			}
			continue
		}
		// other reading → clear whole state
		return nil, false
	}
	if state.nounPos < 0 || state.nounAnalyzedTokenReadings == nil || len(state.nounTokenReadings) == 0 {
		return nil, false
	}
	return state, true
}

// tokenSearchVerbFPSkipPart ports LemmaHelper.tokenSearch(verb.*:[fp], ignore part, FORWARD).
func tokenSearchVerbFPSkipPart(tokens []*languagetool.AnalyzedTokenReadings, pos int) int {
	if tokens == nil || pos < 0 {
		return -1
	}
	for j := pos; j < len(tokens); j++ {
		if tokens[j] == nil {
			continue
		}
		if HasPosTagRE(tokens[j], nounVerbPartRE) {
			continue
		}
		if HasPosTagRE(tokens[j], nounVerbFPVerbRE) {
			return j
		}
	}
	return -1
}
