package uk

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
)

// Java TokenAgreementPrepNounRule.getId()
const TokenAgreementPrepNounRuleID = "UK_PREP_NOUN_INFLECTION_AGREEMENT"

// TokenAgreementPrepNounRule ports prep→noun case government check.
type TokenAgreementPrepNounRule struct {
	*tokenAgreementMatch
	CaseGov *CaseGovernmentHelper
	// Synth optional (Java ukrainian.getSynthesizer()); nil → no suggestions.
	Synth synthesis.Synthesizer
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

		// getExceptionStrong — exception clears prep; skip keeps prep and advances i
		switch ex := GetPrepNounExceptionStrong(tokens, i, prepTok); ex.Type {
		case RuleExceptionException:
			prepPos = -1
			continue
		case RuleExceptionSkip:
			i += ex.Skip
			continue
		}

		if HasPosTagPart(tok, ":v_") {
			// non-normative genitive personal pronouns (їх/його as rod without них form)
			if flag, keep, skipTo := prepPronRodMismatch(tok, clean, tokens, i, prepTok); flag {
				out = append(out, r.newPrepNounMatch(sentence, prepTok, tok, want, ziZnaRemoved, ""))
				prepPos = -1
				continue
			} else if keep {
				continue
			} else if skipTo >= 0 {
				i = skipTo
				continue
			}

			// possessive їх/його/її adj
			if pronAdj := filterReadings(tok, prepNounPronPosRE, prepNounYihLemmas); len(pronAdj) > 0 {
				if !hasVidmPosTagReadings(want, pronAdj) {
					out = append(out, r.newPrepNounMatch(sentence, prepTok, tok, want, ziZnaRemoved, ""))
					prepPos = -1
					continue
				}
				if i < len(tokens)-1 {
					continue // check next noun
				}
			} else if thisLower == "їх" {
				extra := ". Можливо, тут потрібно присвійний займенник «їхній» або нормативна форма р.в. «них»?"
				out = append(out, r.newPrepNounMatch(sentence, prepTok, tok, want, ziZnaRemoved, extra))
				prepPos = -1
				continue
			}

			if HasVidmPosTag(want, tok) {
				prepPos = -1
				continue
			}

			// Java order after hasVidm fails: NonInfl then Infl
			switch ex := GetPrepNounExceptionNonInfl(tokens, i); ex.Type {
			case RuleExceptionException:
				prepPos = -1
				continue
			case RuleExceptionSkip:
				i += ex.Skip
				continue
			}
			switch ex := GetPrepNounExceptionInfl(tokens, prepPos, i); ex.Type {
			case RuleExceptionException:
				prepPos = -1
				continue
			case RuleExceptionSkip:
				i += ex.Skip
				continue
			}

			extra := ""
			if containsStr(want, "v_rod") && prepNounUYuyuRE.MatchString(tok.GetToken()) &&
				HasPosTagRE(tok, prepNounDavMRE) {
				extra = UsedUInsteadOfAMsg
			}
			out = append(out, r.newPrepNounMatch(sentence, prepTok, tok, want, ziZnaRemoved, extra))
		} else {
			// no :v_ — NonInfl may skip (keep prep) or exception (clear); else fall through clear
			switch ex := GetPrepNounExceptionNonInfl(tokens, i); ex.Type {
			case RuleExceptionException:
				prepPos = -1
				continue
			case RuleExceptionSkip:
				i += ex.Skip
				continue
			}
		}
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
// flag=true → emit match; keep=true → continue without clearing (next token);
// skipTo>=0 → jump i to insert end (Java findInsertEnd) without clearing prep.
func prepPronRodMismatch(tok *languagetool.AnalyzedTokenReadings, clean string, tokens []*languagetool.AnalyzedTokenReadings, i int, prepTok *languagetool.AnalyzedTokenReadings) (flag, keep bool, skipTo int) {
	skipTo = -1
	prons := filterReadings(tok, prepNounPronRodRE, prepNounPronLemmas)
	if len(prons) == 0 {
		return false, false, -1
	}
	lower := strings.ToLower(clean)
	if prepNounNihRE.MatchString(lower) {
		return false, false, -1
	}
	if i < len(tokens)-1 && tokens[i+1] != nil {
		next := tokens[i+1]
		if HasPosTagRE(next, regexp.MustCompile(`^(?:noun|adj|adv|part|num|conj:coord|noninfl)`)) ||
			regexp.MustCompile(`^["«„“/$€…]|[a-zA-Z'-]+$`).MatchString(cleanTokenSurface(next)) {
			return false, true, -1
		}
	}
	// Java: insertEndPos = findInsertEnd(prep, tokens, i+1, true)
	if end := findPrepNounInsertEnd(prepTok, tokens, i+1, true); end > 0 {
		return false, false, end
	}
	return true, false, -1
}

// findPrepNounInsertEnd ports TokenAgreementPrepNounRule.findInsertEnd.
// lookForPart is accepted for signature parity (Java never uses it in the body).
func findPrepNounInsertEnd(prepTok *languagetool.AnalyzedTokenReadings, tokens []*languagetool.AnalyzedTokenReadings, i int, lookForPart bool) int {
	_ = prepTok
	_ = lookForPart
	if i >= len(tokens)-2 {
		return -1
	}
	nextPos := i
	tok := tokens[i]
	if tok == nil {
		return -1
	}
	// dead branch in Java (i > length-2 after i >= length-2); kept for twin structure
	if i > len(tokens)-2 {
		return -1
	}
	clean := cleanTokenSurface(tok)
	if regexp.MustCompile(`^же?$`).MatchString(clean) {
		nextPos = i + 1
	}
	if nextPos > len(tokens)-3 {
		if nextPos == i {
			return -1
		}
		return nextPos - 1
	}
	// re-read token at original i for comma/paren unknown-tag inserts
	if tok.IsPosTagUnknown() && regexp.MustCompile(`^[,(]$`).MatchString(clean) {
		commaPos := TokenSearch(tokens, i+1, "", regexp.MustCompile(`^[,)]$`), nil, DirForward)
		if commaPos > i+1 && commaPos < i+6 && commaPos < len(tokens)-1 {
			nextClean := cleanTokenSurface(tokens[commaPos])
			open := strings.ReplaceAll(clean, "(", ")")
			if open == nextClean {
				if tokens[commaPos+1] == nil || cleanTokenSurface(tokens[commaPos+1]) != "що" {
					return commaPos
				}
			}
		}
	}
	if nextPos == i {
		return -1
	}
	return nextPos - 1
}

func prepNounMsg(prep, tok *languagetool.AnalyzedTokenReadings, want []string, ziZnaRemoved bool) string {
	prepTok := ""
	if prep != nil {
		prepTok = prep.GetToken()
	}
	// Java MessageFormat: Прийменник «{0}» вимагає іншого відмінка: {1}, а знайдено: {2}
	reqNames := make([]string, 0, len(want))
	for _, v := range want {
		reqNames = append(reqNames, taguk.VidminokName(v))
	}
	foundNames := foundVidminkyNames(tok)
	msg := "Прийменник «" + prepTok + "» вимагає іншого відмінка: " +
		strings.Join(reqNames, ", ") + ", а знайдено: " + strings.Join(foundNames, ", ")
	if ziZnaRemoved {
		msg += ". Але з.в. вимагається у випадках порівнянн предметів."
	}
	return msg
}

// foundVidminkyNames ports createRuleMatch foundVidminkyNames loop.
func foundVidminkyNames(tok *languagetool.AnalyzedTokenReadings) []string {
	if tok == nil {
		return nil
	}
	var found []string
	seen := map[string]bool{}
	for _, ar := range tok.GetReadings() {
		if ar == nil || ar.GetPOSTag() == nil {
			continue
		}
		pos := *ar.GetPOSTag()
		if !strings.Contains(pos, ":v_") {
			continue
		}
		code := ""
		for _, c := range taguk.VidminkyOrder {
			if strings.Contains(pos, c) {
				code = c
				break
			}
		}
		if code == "" {
			continue
		}
		name := taguk.VidminokName(code)
		if seen[name] {
			// Java: duplicate + :p: → "name (мн.)"
			if strings.Contains(pos, ":p:") {
				name = name + " (мн.)"
				if !seen[name] {
					seen[name] = true
					found = append(found, name)
				}
			}
			continue
		}
		seen[name] = true
		found = append(found, name)
	}
	return found
}

// newPrepNounMatch ports createRuleMatch (message + optional synthesizer suggestions).
func (r *TokenAgreementPrepNounRule) newPrepNounMatch(
	sentence *languagetool.AnalyzedSentence,
	prepTok, tok *languagetool.AnalyzedTokenReadings,
	want []string,
	ziZnaRemoved bool,
	extraMsg string,
) *rules.RuleMatch {
	msg := prepNounMsg(prepTok, tok, want, ziZnaRemoved) + extraMsg
	m := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(), msg)
	m.ShortMessage = r.shortMsg
	if sugs := r.prepNounSuggestions(want, tok); len(sugs) > 0 {
		m.SetSuggestedReplacements(sugs)
	}
	return m
}

// prepNounSuggestions ports createRuleMatch synthesizer loop (v_* remaps).
func (r *TokenAgreementPrepNounRule) prepNounSuggestions(want []string, tok *languagetool.AnalyzedTokenReadings) []string {
	if r == nil || r.Synth == nil || tok == nil || len(want) == 0 {
		return nil
	}
	reqRE := ":(" + strings.Join(want, "|") + ")"
	// Java: append optional or existing :r(in)?anim after case alt
	reqAnim := regexp.MustCompile(`:r(?:in)?anim`)
	seen := map[string]struct{}{}
	var out []string
	for _, ar := range tok.GetReadings() {
		if ar == nil || ar.GetPOSTag() == nil {
			continue
		}
		old := *ar.GetPOSTag()
		apply := reqRE
		if m := reqAnim.FindString(old); m != "" {
			apply += m
		} else {
			// Java: (?:r(in)?anim)? — optional anim; RE2 uses (?:r(?:in)?anim)?
			apply += `(?:r(?:in)?anim)?`
		}
		posTag := regexp.MustCompile(`:v_[a-z]+`).ReplaceAllString(old, apply)
		syn, err := r.Synth.SynthesizeRE(ar, posTag, true)
		if err != nil {
			continue
		}
		for _, s := range syn {
			if s == "" {
				continue
			}
			if _, ok := seen[s]; ok {
				continue
			}
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
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
