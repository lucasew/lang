package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

const TokenAgreementVerbNounRuleID = "UK_VERB_NOUN_INFLECTION_AGREEMENT"

// TokenAgreementVerbNounRule ports TokenAgreementVerbNounRule (state-machine Match).
type TokenAgreementVerbNounRule struct {
	*tokenAgreementMatch
	// CaseGov optional inject; nil → LoadCaseGovernmentHelper().
	CaseGov *CaseGovernmentHelper
}

func NewTokenAgreementVerbNounRule() *TokenAgreementVerbNounRule {
	return NewTokenAgreementVerbNounRuleWithMessages(nil)
}

// NewTokenAgreementVerbNounRuleWithMessages ports the Java ctor (ResourceBundle messages).
func NewTokenAgreementVerbNounRuleWithMessages(messages map[string]string) *TokenAgreementVerbNounRule {
	r := &TokenAgreementVerbNounRule{}
	r.tokenAgreementMatch = &tokenAgreementMatch{
		ruleID: TokenAgreementVerbNounRuleID,
		// Java getDescription / getShort
		description:  "Узгодження дієслова з іменником",
		shortMsg:     "Узгодження дієслова з іменником",
		isLeftToken:  hasVerbOrAdvpReading,
		isRightToken: hasNounAdjNumrObjectReading,
		pairChecker:  r.verbNounAgree,
		exception:    IsVerbNounException,
	}
	initTokenAgreementMeta(r.tokenAgreementMatch, messages)
	return r
}

func (r *TokenAgreementVerbNounRule) caseGov() *CaseGovernmentHelper {
	if r != nil && r.CaseGov != nil {
		return r.CaseGov
	}
	return LoadCaseGovernmentHelper()
}

func (r *TokenAgreementVerbNounRule) verbNounAgree(verb, noun *languagetool.AnalyzedTokenReadings) bool {
	return VerbNounCaseAgree(r.caseGov(), verb, noun)
}

// hasVerbOrAdvpReading ports Java (verb|advp).* master detection.
func hasVerbOrAdvpReading(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	for _, p := range CollectPOSTags(tok) {
		if strings.HasPrefix(p, "verb") || strings.HasPrefix(p, "advp") {
			return true
		}
	}
	return false
}

// hasNounAdjNumrObjectReading ports Java object slots (noun|adj|numr).
func hasNounAdjNumrObjectReading(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	for _, p := range CollectPOSTags(tok) {
		if strings.HasPrefix(p, "noun") || strings.HasPrefix(p, "adj") || strings.HasPrefix(p, "numr") {
			return true
		}
	}
	return false
}

var (
	verbNounArchBadRE = regexp.MustCompile(`.*(arch|bad|slang|alt).*`)
	verbNounInsertRE  = regexp.MustCompile(`^(?:значить|читай|бува|здавалось|здається|здалося)$`)
	verbNounGrosheiRE = regexp.MustCompile(`^(?:грошей|грошенят|дров|товарів|пісень)$`)
	verbNounDashRE    = regexp.MustCompile(`.+ти(?:ся)?-.+ти(?:ся)?`)
)

// Match ports TokenAgreementVerbNounRule.match state machine (not adjacent-pair only).
func (r *TokenAgreementVerbNounRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var out []*rules.RuleMatch
	verbPos := -1
	var verbTok *languagetool.AnalyzedTokenReadings

	// Java assumes tokens[0] is SENT_START; synthetic tests may omit it — start at 0 when so.
	start := 1
	if len(tokens) > 0 && tokens[0] != nil && !tokens[0].IsSentenceStart() && firstPOS(tokens[0]) != "SENT_START" {
		start = 0
	}
	for i := start; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil {
			verbPos = -1
			continue
		}
		if firstPOS(tok) == "" {
			// Java: null posTag0 clears state
			verbPos = -1
			continue
		}

		if hasVerbOrAdvpReading(tok) {
			if IsExceptionVerb(tokens, i) {
				verbPos = -1
				continue
			}
			if IsExceptionVerbSkip(tokens, i) {
				// Java Type.skip: keep existing state, do not install new verb
				continue
			}
			st := getVerbNounState(tokens, i)
			if st < 0 {
				verbPos = -1
				continue
			}
			// Java: state.verbPos == i → wait for object
			verbPos = i
			verbTok = tok
			continue
		}

		if verbPos < 0 || verbTok == nil {
			continue
		}

		if skip := IsVerbNounHardAdjNoun(tokens, i, verbPos); skip >= 0 {
			i += skip
			verbPos = -1
			continue
		}
		if skip := IsVerbNounExceptionSkip(tokens, i); skip >= 0 {
			i += skip
			// Java keeps state
			continue
		}

		// collect object readings (Java nounAdjTokenReadingsVnaz / nounAdjIndirTokenReadings)
		var nazTags, indirTags []string
		hasObj := false
		clearState := false
		for _, p := range CollectPOSTags(tok) {
			if p == "" || strings.HasSuffix(p, "_END") {
				continue
			}
			if strings.HasPrefix(p, "<") {
				clearState = true
				break
			}
			if strings.HasPrefix(p, "noun") || strings.HasPrefix(p, "adj") || strings.HasPrefix(p, "numr") {
				hasObj = true
				if strings.Contains(p, "v_naz") {
					nazTags = append(nazTags, p)
				} else {
					indirTags = append(indirTags, p)
				}
			} else {
				clearState = true
				break
			}
		}
		if clearState || !hasObj {
			verbPos = -1
			continue
		}

		// perform check — Java VerbInflectionHelper overlap on v_naz, then case government on indir
		pass := false
		if len(nazTags) > 0 {
			vInf := GetVerbInflections(CollectPOSTags(verbTok))
			nInf := GetNounInflections(nazTags)
			nInf = append(nInf, GetAdjInflections(nazTags)...)
			if verbInflectionsOverlapLists(vInf, nInf) {
				pass = true
			}
		}

		cases := map[string]struct{}{}
		if !pass && len(indirTags) > 0 {
			cases = r.caseGov().GetCaseGovernmentsFromReadings(verbTok, "verb")
			// dash compound: replaceFirst("(ти(ся)?)-.*", "$1")
			if len(cases) == 0 {
				clean := strings.ToLower(cleanTokenSurface(verbTok))
				if strings.Contains(clean, "-") && verbNounDashRE.MatchString(lemmaOf(verbTok)) {
					cases = caseGovDashVerb(r.caseGov(), verbTok)
				}
			}
			// було ввезено тракторів
			if verbPos > 0 && tokens[verbPos-1] != nil &&
				strings.EqualFold(cleanTokenSurface(tokens[verbPos-1]), "було") &&
				hasPosTagPartVN(verbTok, "impers") {
				cases["v_rod"] = struct{}{}
			}

			tokenLower := strings.ToLower(cleanTokenSurface(tok))
			if _, ok := cases["v_zna"]; ok && verbNounGrosheiRE.MatchString(tokenLower) {
				verbPos = -1
				continue
			}

			// Java: if cases empty OR !hasVidm → leave pass false; else pass = true
			if len(cases) > 0 && hasVidmInTags(cases, indirTags) {
				pass = true
			}
		}

		if !pass {
			// skip pron він/вона/вони v_rod and try next token
			if i < len(tokens)-1 && tokens[i+1] != nil &&
				HasLemmaWithPosRE(tok, []string{"він", "вона", "вони"}, regexp.MustCompile(`noun:.*v_rod.*`)) &&
				HasPosTagRE(tokens[i+1], regexp.MustCompile(`(noun|adj).*`)) {
				continue
			}

			if IsVerbNounException(tokens, verbPos, i) {
				verbPos = -1
				continue
			}

			// Java: flag when hasVidmPosTag(cases, indir) is false.
			// Pure v_naz (empty indir) never flags here — hasVidm on empty list is true.
			if len(nazTags) > 0 || len(indirTags) > 0 {
				if len(cases) == 0 {
					cases = r.caseGov().GetCaseGovernmentsFromReadings(verbTok, "verb")
				}
				if !hasVidmInTags(cases, indirTags) {
					msg := "Не узгоджено дієслово з іменником: \"" + verbTok.GetToken() +
						"\" і \"" + tok.GetToken() + "\""
					m := rules.NewRuleMatch(r, sentence, verbTok.GetStartPos(), tok.GetEndPos(), msg)
					m.ShortMessage = r.shortMsg
					out = append(out, m)
				}
			}
		}
		verbPos = -1
	}
	return out
}

// caseGovDashVerb ports nodash lemma remap for compound verbs like віддати-відрізати.
func caseGovDashVerb(cg *CaseGovernmentHelper, verb *languagetool.AnalyzedTokenReadings) map[string]struct{} {
	out := map[string]struct{}{}
	if cg == nil || verb == nil {
		return out
	}
	dashStrip := regexp.MustCompile(`(ти(?:ся)?)-.*`)
	for _, rdg := range verb.GetReadings() {
		if rdg == nil || rdg.GetLemma() == nil || rdg.GetPOSTag() == nil {
			continue
		}
		if !strings.HasPrefix(*rdg.GetPOSTag(), "verb") {
			continue
		}
		base := dashStrip.ReplaceAllString(*rdg.GetLemma(), "$1")
		for _, c := range cg.GetCaseGovernments(base) {
			out[c] = struct{}{}
		}
	}
	return out
}

// getVerbNounState returns verbPos if token is a valid government master, else -1.
func getVerbNounState(tokens []*languagetool.AnalyzedTokenReadings, i int) int {
	if i < 0 || i >= len(tokens) || tokens[i] == nil {
		return -1
	}
	tok := tokens[i]
	clean := strings.ToLower(cleanTokenSurface(tok))
	if verbNounInsertRE.MatchString(clean) {
		return -1
	}
	// all readings must be verb|advp without abbr; any arch/bad/slang/alt rejects
	for _, p := range CollectPOSTags(tok) {
		if verbNounArchBadRE.MatchString(p) {
			return -1
		}
		if strings.Contains(p, "abbr") {
			return -1
		}
		if !strings.HasPrefix(p, "verb") && !strings.HasPrefix(p, "advp") {
			return -1
		}
	}
	if !hasVerbOrAdvpReading(tok) {
		return -1
	}
	return i
}

func firstPOS(tok *languagetool.AnalyzedTokenReadings) string {
	if tok == nil {
		return ""
	}
	rds := tok.GetReadings()
	if len(rds) == 0 || rds[0] == nil || rds[0].GetPOSTag() == nil {
		return ""
	}
	return *rds[0].GetPOSTag()
}

func cleanTokenSurface(tok *languagetool.AnalyzedTokenReadings) string {
	if tok == nil {
		return ""
	}
	c := tok.GetCleanToken()
	if c == "" {
		c = tok.GetToken()
	}
	return c
}

// hasVidmInTags ports TokenAgreementPrepNounRule.hasVidmPosTag(cases, readings).
// Empty tags (or no :v_ found) → true (incomplete dictionary / pure-v_naz path).
func hasVidmInTags(cases map[string]struct{}, tags []string) bool {
	foundV := false
	for _, p := range tags {
		if p == "" {
			continue
		}
		if strings.Contains(p, ":nv") {
			return true
		}
		if strings.Contains(p, ":v_") {
			foundV = true
			for c := range cases {
				if c != "" && strings.Contains(p, c) {
					return true
				}
			}
		}
	}
	return !foundV
}

func hasPosTagPartVN(tok *languagetool.AnalyzedTokenReadings, part string) bool {
	return HasPosTagPart(tok, part)
}

func verbInflectionsOverlapLists(a, b []VerbInflection) bool {
	for _, x := range a {
		for _, y := range b {
			if x.Equals(y) {
				return true
			}
		}
	}
	return false
}

// VerbNounCaseAgree returns false when a known verb government conflicts with all noun cases.
func VerbNounCaseAgree(cg *CaseGovernmentHelper, verb, noun *languagetool.AnalyzedTokenReadings) bool {
	if verb == nil || noun == nil || cg == nil {
		return true
	}
	var gov []string
	hasGovLemma := false
	for _, r := range verb.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		if !strings.HasPrefix(*r.GetPOSTag(), "verb") && !strings.HasPrefix(*r.GetPOSTag(), "advp") {
			continue
		}
		if r.GetLemma() == nil || *r.GetLemma() == "" {
			continue
		}
		cases := cg.GetCaseGovernments(*r.GetLemma())
		if len(cases) == 0 {
			continue
		}
		hasGovLemma = true
		gov = append(gov, cases...)
	}
	if !hasGovLemma {
		return true
	}
	nounCases := map[string]struct{}{}
	for _, n := range noun.GetReadings() {
		if n == nil || n.GetPOSTag() == nil {
			continue
		}
		pos := *n.GetPOSTag()
		if !strings.HasPrefix(pos, "noun") && !strings.HasPrefix(pos, "adj") && !strings.HasPrefix(pos, "numr") {
			continue
		}
		for _, c := range []string{"v_naz", "v_rod", "v_dav", "v_zna", "v_oru", "v_mis", "v_kly", "v_inf"} {
			if strings.Contains(pos, c) {
				nounCases[c] = struct{}{}
			}
		}
	}
	if len(nounCases) == 0 {
		return true
	}
	hasNonInf := false
	for _, g := range gov {
		if g == "v_inf" {
			continue
		}
		hasNonInf = true
		if _, ok := nounCases[g]; ok {
			return true
		}
	}
	if !hasNonInf {
		return true
	}
	return false
}
