package uk

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
)

// Java TokenAgreementPrepNounRule.getId()
const TokenAgreementPrepNounRuleID = "UK_PREP_NOUN_INFLECTION_AGREEMENT"

// TokenAgreementPrepNounRule ports prep→noun case government check.
type TokenAgreementPrepNounRule struct {
	*tokenAgreementMatch
	CaseGov *CaseGovernmentHelper
}

func hasPrepReading(tok *languagetool.AnalyzedTokenReadings) bool {
	for _, p := range CollectPOSTags(tok) {
		if taguk.IPOSPrep.Match(p) {
			return true
		}
	}
	return false
}

// HasNounOrPronObjectReading treats personal/possessive pronouns as objects for prep government.
func HasNounOrPronObjectReading(tok *languagetool.AnalyzedTokenReadings) bool {
	if HasNounReading(tok) {
		return true
	}
	for _, p := range CollectPOSTags(tok) {
		if strings.Contains(p, "pron") && strings.Contains(p, "v_") {
			return true
		}
	}
	return false
}

func NewTokenAgreementPrepNounRule() *TokenAgreementPrepNounRule {
	return NewTokenAgreementPrepNounRuleWithMessages(nil)
}

// NewTokenAgreementPrepNounRuleWithMessages ports the Java ctor (ResourceBundle messages).
func NewTokenAgreementPrepNounRuleWithMessages(messages map[string]string) *TokenAgreementPrepNounRule {
	cg := LoadCaseGovernmentHelper()
	r := &TokenAgreementPrepNounRule{CaseGov: cg}
	r.tokenAgreementMatch = &tokenAgreementMatch{
		ruleID: TokenAgreementPrepNounRuleID,
		// Java getDescription / getShort
		description:  "Узгодження прийменника та іменника у реченні",
		shortMsg:     "Узгодження прийменника та іменника",
		isLeftToken:  hasPrepReading,
		isRightToken: HasNounOrPronObjectReading,
		pairChecker: func(left, right *languagetool.AnalyzedTokenReadings) bool {
			return prepNounAgree(cg, left, right)
		},
		exception: IsPrepNounException,
	}
	initTokenAgreementMeta(r.tokenAgreementMatch, messages)
	return r
}

// HasVidmPosTag ports TokenAgreementPrepNounRule.hasVidmPosTag.
// posTagsToFind are case substrings like "v_oru"; if no vidminok found on any reading, returns true
// (Java incomplete dictionary path).
func HasVidmPosTag(posTagsToFind []string, tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return true
	}
	rds := tok.GetReadings()
	vidminokFound := false
	for _, token := range rds {
		if token == nil {
			continue
		}
		pos := token.GetPOSTag()
		if pos == nil {
			if len(rds) == 1 {
				return true
			}
			continue
		}
		// Java PosTagHelper.NO_VIDMINOK_SUBSTR
		if strings.Contains(*pos, ":nv") {
			return true
		}
		if strings.Contains(*pos, ":v_") {
			vidminokFound = true
			for _, want := range posTagsToFind {
				if want != "" && strings.Contains(*pos, want) {
					return true
				}
			}
		}
	}
	return !vidminokFound
}

// hasVidmPosTagReadings ports hasVidmPosTag for a subset of AnalyzedToken readings.
func hasVidmPosTagReadings(posTagsToFind []string, readings []*languagetool.AnalyzedToken) bool {
	vidminokFound := false
	for _, token := range readings {
		if token == nil {
			continue
		}
		pos := token.GetPOSTag()
		if pos == nil {
			if len(readings) == 1 {
				return true
			}
			continue
		}
		if strings.Contains(*pos, ":nv") {
			return true
		}
		if strings.Contains(*pos, ":v_") {
			vidminokFound = true
			for _, want := range posTagsToFind {
				if want != "" && strings.Contains(*pos, want) {
					return true
				}
			}
		}
	}
	return !vidminokFound
}

func prepNounAgree(cg *CaseGovernmentHelper, prep, noun *languagetool.AnalyzedTokenReadings) bool {
	if cg == nil || prep == nil || noun == nil {
		return true
	}
	// lemma from prep token surface / lemma
	lemma := prep.GetToken()
	// strip soft hyphen / combining marks from surface lemma
	lemma = CleanIgnoreChars(lemma)
	for _, r := range prep.GetReadings() {
		if r != nil && r.GetLemma() != nil && *r.GetLemma() != "" {
			lemma = CleanIgnoreChars(*r.GetLemma())
			break
		}
	}
	govs := cg.GetCaseGovernments(lemma)
	if len(govs) == 0 {
		return true // unknown prep — no flag
	}
	nounInfs := GetNounCaseInflections(CollectPOSTags(noun))
	if len(nounInfs) == 0 {
		// try free case scan for pron tags
		for _, p := range CollectPOSTags(noun) {
			for _, c := range []string{"v_naz", "v_rod", "v_dav", "v_zna", "v_oru", "v_mis", "v_kly"} {
				if strings.Contains(p, c) && cg.HasCaseGovernment(lemma, c) {
					return true
				}
			}
		}
		return true // insufficient
	}
	for _, inf := range nounInfs {
		if cg.HasCaseGovernment(lemma, inf.Case) {
			return true
		}
	}
	return false
}

var (
	prepNounQuotes = map[string]struct{}{
		"«": {}, "\"": {}, "„": {}, "“": {},
	}
	prepNounZZI   = map[string]struct{}{"з": {}, "зі": {}, "із": {}}
	prepNounZZIZO = map[string]struct{}{"з": {}, "зі": {}, "із": {}, "зо": {}}
	prepNounNull  = map[string]struct{}{
		"шляхом": {}, "од": {}, "поруч": {}, "ради": {},
	}
	prepNounPronRodRE  = regexp.MustCompile(`noun:unanim:.:v_rod.*pron.*`)
	prepNounPronPosRE  = regexp.MustCompile(`adj.*pron:pos`) // RE2: exclude :bad in filterReadings
	prepNounDavMRE     = regexp.MustCompile(`noun.*?:m:v_dav.*`)
	prepNounAnimNazRE  = regexp.MustCompile(`noun:anim:.:v_naz.*`)
	prepNounUYuyuRE    = regexp.MustCompile(`.*[ую]$`)
	prepNounLatinAfter = regexp.MustCompile(`.*[а-яіїєґ0-9]$`)
	prepNounPoverhRE   = regexp.MustCompile(`.*поверх(ов|ів).*`)
	prepNounZnaNumRE   = regexp.MustCompile(`noun:inanim:[fnm]:v_zna.*num.*|^num`)
	prepNounZnaLemmaRE = regexp.MustCompile(`noun:inanim:[mnf]:v_zna.*`)
	prepNounAdjZnaRE   = regexp.MustCompile(`adj:[mnf]:v_zna.*`)
	prepNounApproxTag  = regexp.MustCompile(`noun.*v_oru.*|^adv|^part`)
	prepNounPronLemmas = map[string]struct{}{
		"вони": {}, "він": {}, "вона": {}, "воно": {},
	}
	prepNounYihLemmas = map[string]struct{}{
		"їх": {}, "його": {}, "її": {},
	}
	prepNounApproxLemmas = []string{
		"розмір", "величина", "товщина", "вартість", "ріст", "зріст", "висота", "глибина", "діаметр", "вага", "обсяг", "площа",
		"приблизно", "десь", "завбільшки", "завширшки", "завдовжки", "завтовшки", "заввишки", "завглибшки",
	}
	prepNounNihRE = regexp.MustCompile(`^(?:них|нього|неї)(?:-[а-я]+)?$`)
)

// Match ports TokenAgreementPrepNounRule.match state machine.
func (r *TokenAgreementPrepNounRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	cg := r.CaseGov
	if cg == nil {
		cg = LoadCaseGovernmentHelper()
	}
	var out []*rules.RuleMatch
	prepPos := -1
	var prepTok *languagetool.AnalyzedTokenReadings
	ziZnaRemoved := false

	start := 1
	if len(tokens) > 0 && tokens[0] != nil && !tokens[0].IsSentenceStart() && firstPOS(tokens[0]) != "SENT_START" {
		start = 0
	}

	for i := start; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil {
			prepPos = -1
			continue
		}
		clean := cleanTokenSurface(tok)
		if _, q := prepNounQuotes[clean]; q {
			continue
		}
		if firstPOS(tok) == "" {
			prepPos = -1
			continue
		}

		// single uppercase Latin/Cyr letter after Cyrillic (гепатит В)
		if i > 0 && utf8.RuneCountInString(clean) == 1 {
			r0, _ := utf8.DecodeRuneInString(clean)
			if unicode.IsUpper(r0) && tok.IsWhitespaceBefore() && tokens[i-1] != nil {
				prev := tokens[i-1].GetToken()
				if prepNounLatinAfter.MatchString(prev) {
					prepPos = -1
					continue
				}
			}
		}

		// multiword tags starting with <
		if mw := getPrepMultiwordToken(tok); mw != nil {
			mwLem := ""
			if mw.GetLemma() != nil {
				mwLem = *mw.GetLemma()
			}
			mwPos := ""
			if mw.GetPOSTag() != nil {
				mwPos = *mw.GetPOSTag()
			}
			lower := strings.ToLower(clean)
			if _, ok := prepNounZZI[lower]; ok && strings.HasPrefix(mwLem, "згідно ") {
				prepPos = i
				prepTok = tok
				ziZnaRemoved = false
				continue
			}
			if strings.HasPrefix(firstPOS(tok), "prep") {
				prepPos = -1
				continue
			}
			if !strings.Contains(mwPos, "adv") && !strings.Contains(mwPos, "insert") {
				prepPos = -1
			}
			continue
		}

		if strings.HasPrefix(firstPOS(tok), "prep") || hasPrepReading(tok) {
			prep := strings.ToLower(clean)
			if prep == "понад" {
				continue // keep prior state
			}
			if _, kill := prepNounNull[prep]; kill {
				prepPos = -1
				continue
			}
			prepPos = i
			prepTok = tok
			ziZnaRemoved = false
			continue
		}

		if prepPos < 0 || prepTok == nil {
			continue
		}

		thisLower := strings.ToLower(clean)
		if thisLower == "ван" || clean == "Фон" {
			continue
		}
		if thisLower == "та" {
			prepPos = -1
			continue
		}

		// expected cases
		posTagsToFind := map[string]struct{}{}
		prepLemma := prepLemmaOf(prepTok)

		if prepLemma == "замість" {
			posTagsToFind["v_naz"] = struct{}{}
		} else if prepLemma == "за" {
			if prepPos > 0 && tokens[prepPos-1] != nil &&
				strings.EqualFold(cleanTokenSurface(tokens[prepPos-1]), "що") {
				posTagsToFind["v_naz"] = struct{}{}
			}
		}

		// quoted titles
		if i > 0 && tokens[i-1] != nil {
			if _, q := prepNounQuotes[cleanTokenSurface(tokens[i-1])]; q {
				if IsCapitalized(clean) || strings.EqualFold(cleanTokenSurface(prepTok), "замість") {
					prepPos = -1
					continue
				}
				posTagsToFind["v_naz"] = struct{}{}
			}
		}

		expected := cg.GetCaseGovernmentsFromReadings(prepTok, "prep")
		if _, isZ := prepNounZZIZO[prepLemma]; isZ {
			if strings.EqualFold(clean, "нізвідки") {
				prepPos = -1
				continue
			}
			if _, isZZI := prepNounZZI[prepLemma]; isZZI &&
				i >= 2 && tokens[i-2] != nil &&
				strings.EqualFold(cleanTokenSurface(tokens[i-2]), "згідно") {
				expected = map[string]struct{}{"v_oru": {}}
			} else if !isLikelyApproxWithZi(tokens, i, prepPos) {
				delete(expected, "v_zna")
				ziZnaRemoved = true
			}
		}
		delete(expected, "v_inf")
		for c := range expected {
			posTagsToFind[c] = struct{}{}
		}
		want := mapKeys(posTagsToFind)

		// strong exception via combined helper
		if IsPrepNounException(tokens, prepPos, i) {
			// Java strong may clear or skip; combined helper is soft exception → clear
			// But "не" skip next should not clear like Java Type.skip — approximate: clear for now
			// only if pure exception arms; skip-style "не" returns true too
			if clean == "не" && i < len(tokens)-1 && HasPosTagStart(tokens[i+1], "ad") {
				continue // keep state over "не"
			}
			prepPos = -1
			continue
		}

		if HasPosTagPart(tok, ":v_") {
			// non-normative genitive personal pronouns (їх/його as rod without них form)
			if flag, keep := prepPronRodMismatch(tok, clean, tokens, i, want); flag {
				msg := prepNounMsg(prepTok, want, ziZnaRemoved)
				m := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(), msg)
				m.ShortMessage = r.shortMsg
				out = append(out, m)
				prepPos = -1
				continue
			} else if keep {
				continue
			}

			// possessive їх/його/її adj
			if pronAdj := filterReadings(tok, prepNounPronPosRE, prepNounYihLemmas); len(pronAdj) > 0 {
				if !hasVidmPosTagReadings(want, pronAdj) {
					msg := prepNounMsg(prepTok, want, ziZnaRemoved)
					m := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(), msg)
					m.ShortMessage = r.shortMsg
					out = append(out, m)
					prepPos = -1
					continue
				}
				if i < len(tokens)-1 {
					continue // check next noun
				}
			} else if thisLower == "їх" {
				msg := prepNounMsg(prepTok, want, ziZnaRemoved)
				msg += ". Можливо, тут потрібно присвійний займенник «їхній» або нормативна форма р.в. «них»?"
				m := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(), msg)
				m.ShortMessage = r.shortMsg
				out = append(out, m)
				prepPos = -1
				continue
			}

			if HasVidmPosTag(want, tok) {
				prepPos = -1
				continue
			}

			// infl/non-infl exceptions already partially covered; re-check pair helper
			if IsPrepNounException(tokens, prepPos, i) {
				prepPos = -1
				continue
			}

			msg := prepNounMsg(prepTok, want, ziZnaRemoved)
			if containsStr(want, "v_rod") && prepNounUYuyuRE.MatchString(tok.GetToken()) &&
				HasPosTagRE(tok, prepNounDavMRE) {
				msg += UsedUInsteadOfAMsg
			}
			m := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(), msg)
			m.ShortMessage = r.shortMsg
			out = append(out, m)
		}
		// no :v_ — non-infl exception may apply; then always clear (Java)
		prepPos = -1
	}
	return out
}

func prepLemmaOf(prep *languagetool.AnalyzedTokenReadings) string {
	if prep == nil {
		return ""
	}
	rds := prep.GetReadings()
	if len(rds) > 0 && rds[0] != nil && rds[0].GetLemma() != nil {
		return strings.ToLower(CleanIgnoreChars(*rds[0].GetLemma()))
	}
	return strings.ToLower(CleanIgnoreChars(cleanTokenSurface(prep)))
}

func getPrepMultiwordToken(tok *languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedToken {
	if tok == nil {
		return nil
	}
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		if strings.HasPrefix(*r.GetPOSTag(), "<") {
			return r
		}
	}
	return nil
}

func mapKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func filterReadings(tok *languagetool.AnalyzedTokenReadings, posRE *regexp.Regexp, lemmas map[string]struct{}) []*languagetool.AnalyzedToken {
	var out []*languagetool.AnalyzedToken
	if tok == nil {
		return out
	}
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		pos := *r.GetPOSTag()
		if !posRE.MatchString(pos) {
			continue
		}
		// Java: adj.*pron:pos(?!:bad).* — drop bad possessive tags
		if strings.Contains(pos, "pron:pos") && strings.Contains(pos, ":bad") {
			continue
		}
		if r.GetLemma() == nil {
			continue
		}
		if _, ok := lemmas[*r.GetLemma()]; ok {
			out = append(out, r)
		}
	}
	return out
}

// prepPronRodMismatch ports the non-normative rod personal-pronoun arm.
// flag=true → emit match; keep=true → continue without clearing (next token).
func prepPronRodMismatch(tok *languagetool.AnalyzedTokenReadings, clean string, tokens []*languagetool.AnalyzedTokenReadings, i int, _ []string) (flag, keep bool) {
	prons := filterReadings(tok, prepNounPronRodRE, prepNounPronLemmas)
	if len(prons) == 0 {
		return false, false
	}
	lower := strings.ToLower(clean)
	if prepNounNihRE.MatchString(lower) {
		return false, false
	}
	if i < len(tokens)-1 && tokens[i+1] != nil {
		next := tokens[i+1]
		if HasPosTagRE(next, regexp.MustCompile(`^(?:noun|adj|adv|part|num|conj:coord|noninfl)`)) ||
			regexp.MustCompile(`^["«„“/$€…]|[a-zA-Z'-]+$`).MatchString(cleanTokenSurface(next)) {
			return false, true
		}
	}
	return true, false
}

func prepNounMsg(prep *languagetool.AnalyzedTokenReadings, want []string, ziZnaRemoved bool) string {
	prepTok := ""
	if prep != nil {
		prepTok = prep.GetToken()
	}
	msg := "Прийменник «" + prepTok + "» вимагає іншого відмінка"
	if len(want) > 0 {
		msg += ": " + strings.Join(want, ", ")
	}
	if ziZnaRemoved {
		msg += ". Але з.в. вимагається у випадках порівнянн предметів."
	}
	return msg
}

// isLikelyApproxWithZi ports TokenAgreementPrepNounRule.isLikelyApproxWithZi.
func isLikelyApproxWithZi(tokens []*languagetool.AnalyzedTokenReadings, i, prepPos int) bool {
	if i < 0 || i >= len(tokens) || tokens[i] == nil {
		return false
	}
	if prepNounPoverhRE.MatchString(cleanTokenSurface(tokens[i])) {
		return true
	}
	if HasPosTagRE(tokens[i], prepNounZnaNumRE) {
		return true
	}
	// TIME/DISTANCE/PSEUDO_NUM + spoon lemmas on v_zna
	approxLemmas := append([]string{}, prepNounApproxLemmas...)
	approxLemmas = append(approxLemmas, DistanceLemmas...)
	approxLemmas = append(approxLemmas, PseudoNumLemmas...)
	approxLemmas = append(approxLemmas, "ложка", "ложечка")
	// TimePlusLemmas is a set — expand keys
	for s := range TimePlusLemmas {
		approxLemmas = append(approxLemmas, s)
	}
	if HasLemmaWithPosRE(tokens[i], approxLemmas, prepNounZnaLemmaRE) {
		return true
	}
	if i < len(tokens)-1 && HasPosTagRE(tokens[i], prepNounAdjZnaRE) &&
		HasLemmaWithPosRE(tokens[i+1], approxLemmas, prepNounZnaLemmaRE) {
		return true
	}
	if prepPos > 0 && tokens[prepPos-1] != nil &&
		HasLemmaWithPosRE(tokens[prepPos-1], prepNounApproxLemmas, prepNounApproxTag) {
		return true
	}
	if i < len(tokens)-1 && tokens[i+1] != nil &&
		HasLemmaWithPosRE(tokens[i+1], prepNounApproxLemmas, prepNounApproxTag) {
		return true
	}
	return false
}
