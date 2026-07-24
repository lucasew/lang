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

// TokenAgreementNounVerbRule ports TokenAgreementNounVerbRule (state-machine Match).
type TokenAgreementNounVerbRule struct {
	*tokenAgreementMatch
}

func hasVerbReading(tok *languagetool.AnalyzedTokenReadings) bool {
	for _, p := range CollectPOSTags(tok) {
		if taguk.IPOSVerb.Match(p) {
			return true
		}
	}
	return false
}

func NewTokenAgreementNounVerbRule() *TokenAgreementNounVerbRule {
	return NewTokenAgreementNounVerbRuleWithMessages(nil)
}

// NewTokenAgreementNounVerbRuleWithMessages ports the Java ctor (ResourceBundle messages).
func NewTokenAgreementNounVerbRuleWithMessages(messages map[string]string) *TokenAgreementNounVerbRule {
	r := &TokenAgreementNounVerbRule{}
	r.tokenAgreementMatch = &tokenAgreementMatch{
		ruleID: TokenAgreementNounVerbRuleID,
		// Java getDescription / getShort
		description:  "Узгодження іменника та дієслова за родом, числом та особою",
		shortMsg:     "Узгодження іменника з дієсловом",
		isLeftToken:  HasNounOrPronSubjectReading,
		isRightToken: hasVerbReading,
		pairChecker:  nounVerbAgree,
		exception:    IsNounVerbException,
	}
	initTokenAgreementMeta(r.tokenAgreementMatch, messages)
	return r
}

func nounVerbAgree(noun, verb *languagetool.AnalyzedTokenReadings) bool {
	nTags := CollectPOSTags(noun)
	vTags := CollectPOSTags(verb)
	// proper names soft: prop without clear person/number often skip
	if isProperNameOnly(nTags) && len(GetNounInflections(nTags)) == 0 {
		return true
	}
	// personal pronouns: use person/number soft matrix
	if hasPronPers(nTags) {
		return pronVerbAgree(nTags, vTags)
	}
	if len(GetNounInflections(nTags)) == 0 || len(GetVerbInflections(vTags)) == 0 {
		return true // insufficient data
	}
	return VerbInflectionsOverlap(vTags, nTags)
}

func hasPronPers(tags []string) bool {
	for _, t := range tags {
		if strings.Contains(t, "pron:pers") {
			return true
		}
	}
	return false
}

func isProperNameOnly(tags []string) bool {
	if len(tags) == 0 {
		return false
	}
	for _, t := range tags {
		if !strings.Contains(t, "prop") {
			return false
		}
	}
	return true
}

// pronVerbAgree soft-matches personal pronouns to verb person/number.
func pronVerbAgree(nTags, vTags []string) bool {
	// extract :1/:2/:3 and s/p from both
	var nPers, nNum, vPers, vNum string
	for _, t := range nTags {
		if !strings.Contains(t, "pron:pers") {
			continue
		}
		for _, p := range []string{":1", ":2", ":3"} {
			if strings.Contains(t, p) {
				nPers = p
			}
		}
		if strings.Contains(t, ":p:") || strings.HasSuffix(t, ":p") || strings.Contains(t, ":p:v_") {
			nNum = "p"
		} else if strings.Contains(t, ":s:") || strings.Contains(t, ":m:") || strings.Contains(t, ":f:") || strings.Contains(t, ":n:") {
			nNum = "s"
		}
		// Ukrainian: noun:…:p:v_naz:pron:pers:1
		if strings.Contains(t, ":p:") {
			nNum = "p"
		}
	}
	for _, t := range vTags {
		if !strings.HasPrefix(t, "verb") {
			continue
		}
		for _, p := range []string{":1", ":2", ":3"} {
			if strings.Contains(t, p) {
				vPers = p
			}
		}
		if strings.Contains(t, ":p:") || strings.Contains(t, ":p:3") || strings.Contains(t, "past:p") {
			vNum = "p"
		} else if strings.Contains(t, ":s:") || strings.Contains(t, "past:m") || strings.Contains(t, "past:f") || strings.Contains(t, "past:n") {
			vNum = "s"
		}
	}
	if nPers == "" || vPers == "" {
		return true // insufficient
	}
	if nPers != vPers {
		return false
	}
	if nNum != "" && vNum != "" && nNum != vNum {
		return false
	}
	return true
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

// isPredictOrInsert ports PosTagHelper.isPredictOrInsert.
func isPredictOrInsertPOS(pos string) bool {
	return nounVerbPredInsRE.MatchString(pos)
}

// Match ports TokenAgreementNounVerbRule.match state machine.
func (r *TokenAgreementNounVerbRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var out []*rules.RuleMatch

	nounPos := -1
	var nounTok *languagetool.AnalyzedTokenReadings
	var nounTags []string // v_naz noun / яка readings used for master inflections

	start := 1
	if len(tokens) > 0 && tokens[0] != nil && !tokens[0].IsSentenceStart() && firstPOS(tokens[0]) != "SENT_START" {
		start = 0
	}
	for i := start; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil {
			nounPos = -1
			continue
		}
		if firstPOS(tok) == "" {
			nounPos = -1
			continue
		}

		clean := cleanTokenSurface(tok)
		// subject start: noun.*:v_naz or "яка"
		if HasPosTagRE(tok, nounVerbNazRE) || strings.EqualFold(clean, "яка") {
			nPos, nTok, nTags, ok := collectNounVerbSubject(tokens, i)
			if !ok {
				nounPos = -1
				continue
			}
			nounPos = nPos
			nounTok = nTok
			nounTags = nTags
			continue
		}

		if nounPos < 0 || nounTok == nil {
			continue
		}

		// ignorable particles / pure adv (Java continue — keep state)
		if _, ok := nounVerbSkipToks[tok.GetToken()]; ok {
			continue
		}
		if hasPosTagPartAll(tok, "adv") {
			continue
		}

		// collect verb readings on this token
		var verbTags []string
		clear := false
		for _, p := range CollectPOSTags(tok) {
			if p == "" || p == "SENT_END" || p == "PARA_END" {
				continue
			}
			if strings.HasPrefix(p, "<") {
				clear = true
				break
			}
			if strings.HasPrefix(p, "verb") {
				verbTags = append(verbTags, p)
			} else if isPredictOrInsertPOS(p) {
				// ignore
			} else {
				clear = true
				break
			}
		}
		if clear || len(verbTags) == 0 {
			nounPos = -1
			continue
		}

		master := GetNounInflections(nounTags)
		slave := GetVerbInflections(verbTags)
		if !verbInflectionsOverlapLists(master, slave) {
			// Java: clear subject + break entire match loop
			if IsNounVerbException(tokens, nounPos, i) {
				break
			}
			// Java flags whenever Collections.disjoint — including empty master with non-empty slave
			kind := "іменник"
			if HasLemmaToken(nounTok, "який") {
				kind = "займенник"
			}
			// Java: "Не узгоджено %s з дієсловом: \"%s\" (%s) і \"%s\" (%s)"
			msg := "Не узгоджено " + kind + " з дієсловом: \"" + nounTok.GetToken() +
				"\" (" + formatVerbPersonInflections(master, true) + ") і \"" + tok.GetToken() +
				"\" (" + formatVerbPersonInflections(slave, false) + ")"
			m := rules.NewRuleMatch(r, sentence, nounTok.GetStartPos(), tok.GetEndPos(), msg)
			m.ShortMessage = r.shortMsg
			out = append(out, m)
		}
		nounPos = -1
	}
	return out
}

// formatVerbPersonInflections ports TokenAgreementNounVerbRule.formatInflections.
// noun flag is unused in Java body but kept for signature parity.
func formatVerbPersonInflections(infs []VerbInflection, noun bool) string {
	_ = noun
	if len(infs) == 0 {
		return ""
	}
	// Java Collections.sort by gender GEN_ORDER
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
	// InflectionHelper.GEN_ORDER; null gender → 0 in Java compareTo
	if inf.Gender == "" {
		return 0
	}
	if o, ok := genOrder[inf.Gender]; ok {
		return o
	}
	return 99
}

// collectNounVerbSubject ports the Java noun/яка state builder. ok=false → clear state.
func collectNounVerbSubject(tokens []*languagetool.AnalyzedTokenReadings, i int) (nounPos int, nounTok *languagetool.AnalyzedTokenReadings, nounTags []string, ok bool) {
	if i < 0 || i >= len(tokens) || tokens[i] == nil {
		return -1, nil, nil, false
	}
	tok := tokens[i]
	clean := cleanTokenSurface(tok)
	nounPos = -1
	for _, rdg := range tok.GetReadings() {
		if rdg == nil || rdg.GetPOSTag() == nil {
			continue
		}
		pos := *rdg.GetPOSTag()
		lem := ""
		if rdg.GetLemma() != nil {
			lem = *rdg.GetLemma()
		}
		if lem == "який" && strings.Contains(pos, ":f:v_naz") {
			nounPos = i
			nounTok = tok
			nounTags = append(nounTags, pos)
			continue
		}
		if strings.EqualFold(clean, "хто") {
			// ignore: хто + future f/p verb
			if tokenSearchVerbFPSkipPart(tokens, i+1) >= 0 {
				return -1, nil, nil, false
			}
			if i < len(tokens)-1 && hasPosTagPartAll(tokens[i+1], "adv") &&
				tokenSearchVerbFPSkipPart(tokens, i+2) >= 0 {
				return -1, nil, nil, false
			}
		}
		if strings.HasPrefix(pos, "noun") && strings.Contains(pos, "v_naz") {
			nounPos = i
			nounTok = tok
			nounTags = append(nounTags, pos)
			continue
		}
		if strings.HasPrefix(pos, "noun") && strings.Contains(pos, "v_kly") {
			continue // ignore
		}
		if isPredictOrInsertPOS(pos) {
			continue
		}
		// Java: adj:.:(v_naz|v_kly).* || (adj:m:v_zna:rinanim && !prep) && !blocked
		// (blocked list binds only to the rinanim branch via && precedence)
		if nounVerbAdjNazRE.MatchString(pos) {
			continue // adj readings not used for master inflections
		}
		if nounVerbAdjZnaRE.MatchString(pos) &&
			!(i > 0 && tokens[i-1] != nil && HasPosTagStart(tokens[i-1], "prep")) {
			lower := strings.ToLower(tok.GetToken())
			if _, block := nounVerbAdjBlock[lower]; block {
				return -1, nil, nil, false
			}
			continue
		}
		// other reading → clear whole state
		return -1, nil, nil, false
	}
	if nounPos < 0 || nounTok == nil || len(nounTags) == 0 {
		return -1, nil, nil, false
	}
	return nounPos, nounTok, nounTags, true
}

// tokenSearchVerbFPSkipPart ports LemmaHelper.tokenSearch(verb.*:[fp], ignore part, FORWARD).
func tokenSearchVerbFPSkipPart(tokens []*languagetool.AnalyzedTokenReadings, pos int) int {
	if tokens == nil || pos < 0 {
		return -1
	}
	for i := pos; i < len(tokens); i++ {
		if tokens[i] == nil {
			continue
		}
		if HasPosTagRE(tokens[i], nounVerbPartRE) {
			continue
		}
		if HasPosTagRE(tokens[i], nounVerbFPVerbRE) {
			return i
		}
	}
	return -1
}
