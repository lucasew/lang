package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
)

const TokenAgreementNumrNounRuleID = "UK_NUMR_NOUN_INFLECTION_AGREEMENT"

// TokenAgreementNumrNounRule ports TokenAgreementNumrNounRule.
type TokenAgreementNumrNounRule struct {
	*tokenAgreementMatch
}

func hasNumrReading(tok *languagetool.AnalyzedTokenReadings) bool {
	for _, p := range CollectPOSTags(tok) {
		if taguk.IPOSNumr.Match(p) || taguk.IPOSNumber.Match(p) {
			return true
		}
	}
	return false
}

// nounForcePattern ports TokenAgreementNumrNounRule.NOUN_FORCE_PATTERN (Java Matcher.matches).
// Full-string only — do not invent extra plurals like "тони" beyond the Java regex.
var nounForcePattern = regexp.MustCompile(
	`^(?:чоловік|солдат|тон|(?:нано|мікро|мілі|дека|кіло|мега|гіга|тера|пета)?(?:герц|байт|біт|бар|бер|ват|вольт|децибел|рентген|моль|мікрон|грам|аршин|лат|карат))$`,
)

// Java TokenAgreementNumrNounRule surface patterns for fractional / half numerals
// (Matcher.matches / String.matches — full-string).
var (
	// numrToken.matches("(один-|одне-)?півтора")
	numrPivtoraRE = regexp.MustCompile(`^(?:один-|одне-)?півтора$`)
	// numrToken.matches("(одн.+-)?півтори")
	numrPivtoryRE = regexp.MustCompile(`^(?:одн.+-)?півтори$`)
	// numrToken.matches("пів")
	numrPivRE = regexp.MustCompile(`^пів$`)
	// _FRACT = Pattern.compile(".*,[1-9]+")
	numrFractRE = regexp.MustCompile(`,[1-9]+$`)

	numrNounIgnoreRE  = regexp.MustCompile(`(?:prop|noun.*pron|v_oru)`)
	numrNounNumrAllRE = regexp.MustCompile(`^noun:inanim:(?:[mf]:v_naz|p:v_(?:naz|rod)).*:numr.*|^numr.*abbr.*|^number$`)
	numrDashLetterRE  = regexp.MustCompile(`.*[0-9]-[а-яіїєґ].*`)
	numrHalfPrepRE    = regexp.MustCompile(`^(?:з|із|зі)$`)
	numrHalfNounRE    = regexp.MustCompile(`^(?:половиною|третиною|чвертю|гаком)$`)
	numrAdjPRodNazRE  = regexp.MustCompile(`^adj:p:v_(?:rod|naz).*`)
	numrMRodAyaRE     = regexp.MustCompile(`.*:m:v_rod.*`)
	numrAyaTokenRE    = regexp.MustCompile(`.*[ая]$`)
	numr1_5RE         = regexp.MustCompile(`^(?:[0-9]+[–-])?1,5$`)
	numr5_5RE         = regexp.MustCompile(`(?:[0-9]+[–-])?(?:[0-9 ]*[05-9]|[0-9 ]*1[1-4]),5$`)
	numr5to9RE        = regexp.MustCompile(`[0-9 ]*(?:[5-90]|1[2-4])$`)
	numr5to9AlphaRE   = regexp.MustCompile(`^(?:(?:.+-)?(?:п.ять|шість|сім|вісім|(?:три)?дев.?ять|.*дцять|сорок|.*десять?|дев.яносто|сто|двісті|триста|чотириста|півтораста|.+сот)|(?:де)?кілька|кількох|аніскільки)$`)
	numrDvoeEtcRE     = regexp.MustCompile(`^(?:(?:.+-)?(?:двоє|двох|троє|.+еро|.+ьох)|обидвоє|обидвох|обоє|обох|двійко)$`)
	numrBagatoRE      = regexp.MustCompile(`^(?:(?:не)?багато|багато-багато|(?:не|чи)?мало|с[тк]ільки(?:-то|сь)?|.+-скільки|кілько)$`)
	numrRazRE         = regexp.MustCompile(`^(?:раз|рази|разу|разів)$`)
	numrDesyatykhRE   = regexp.MustCompile(`^(?:десятих|сотих|тисячних|третіх|четвертих)$`)
	numrNynRodRE      = regexp.MustCompile(`^noun:anim:m:v_rod.*`)
	numrNynNazRE      = regexp.MustCompile(`^noun:anim:p:v_naz.*`)
	numrNynRodTokenRE = regexp.MustCompile(`.*нин[ая]$`)
	numrNynNazTokenRE = regexp.MustCompile(`.*ни$`)
	numrInanimPZnaRE  = regexp.MustCompile(`^noun:inanim:p:v_zna.*`)
	numrAdjPZnaRE     = regexp.MustCompile(`^adj:p:v_zna.*`)
	numrNounPRodRE    = regexp.MustCompile(`^noun:.*p:v_rod.*`)
	numrDavMRE        = regexp.MustCompile(`noun.*?:m:v_dav.*`)
)

func NewTokenAgreementNumrNounRule() *TokenAgreementNumrNounRule {
	return NewTokenAgreementNumrNounRuleWithMessages(nil)
}

// NewTokenAgreementNumrNounRuleWithMessages ports the Java ctor (ResourceBundle messages).
func NewTokenAgreementNumrNounRuleWithMessages(messages map[string]string) *TokenAgreementNumrNounRule {
	r := &TokenAgreementNumrNounRule{}
	r.tokenAgreementMatch = &tokenAgreementMatch{
		ruleID: TokenAgreementNumrNounRuleID,
		// Java getDescription / getShort
		description:  "Узгодження відмінків, роду і числа числівника та іменника",
		shortMsg:     "Узгодження числівника та іменника",
		isLeftToken:  hasNumrReading,
		isRightToken: HasNounReading,
		pairChecker: func(left, right *languagetool.AnalyzedTokenReadings) bool {
			if IsForceNounException(left, right) {
				return true
			}
			if IsFractionalNumrException(left, right) {
				return true
			}
			return NumrNounAgree(CollectPOSTags(left), CollectPOSTags(right))
		},
		exception: IsNumrNounException,
	}
	initTokenAgreementMeta(r.tokenAgreementMatch, messages)
	return r
}

// IsForceNounException ports Java NOUN_FORCE_PATTERN.matcher(cleanTokenLower).matches().
func IsForceNounException(numr, noun *languagetool.AnalyzedTokenReadings) bool {
	if noun == nil {
		return false
	}
	// Java uses getCleanToken().toLowerCase() on the noun surface.
	clean := strings.ToLower(noun.GetCleanToken())
	if clean == "" {
		clean = strings.ToLower(noun.GetToken())
	}
	return nounForcePattern.MatchString(clean)
}

// IsFractionalNumrException ports Java half/fractional numeral surfaces.
func IsFractionalNumrException(numr, noun *languagetool.AnalyzedTokenReadings) bool {
	if numr == nil {
		return false
	}
	tok := strings.ToLower(numr.GetToken())
	if numrPivtoraRE.MatchString(tok) || numrPivtoryRE.MatchString(tok) ||
		numrPivRE.MatchString(tok) || numrFractRE.MatchString(tok) {
		return true
	}
	clean := strings.ToLower(numr.GetCleanToken())
	if clean != "" && clean != tok {
		if numrPivtoraRE.MatchString(clean) || numrPivtoryRE.MatchString(clean) ||
			numrPivRE.MatchString(clean) || numrFractRE.MatchString(clean) {
			return true
		}
	}
	return false
}

// hasNumrPOS ports NUMR_PATTERN numr(?!.*abbr).*
func hasNumrPOS(tok *languagetool.AnalyzedTokenReadings) bool {
	for _, p := range CollectPOSTags(tok) {
		if strings.HasPrefix(p, "numr") && !strings.Contains(p, "abbr") {
			return true
		}
	}
	return false
}

// Match ports TokenAgreementNumrNounRule.match state machine (core arms; no synthesizer).
func (r *TokenAgreementNumrNounRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var out []*rules.RuleMatch

	numrPos := -1
	var numrTok *languagetool.AnalyzedTokenReadings
	var numrTags []string
	isNumber := false

	start := 1
	if len(tokens) > 0 && tokens[0] != nil && !tokens[0].IsSentenceStart() && firstPOS(tokens[0]) != "SENT_START" {
		start = 0
	}

	for i := start; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil {
			numrPos = -1
			continue
		}
		if firstPOS(tok) == "" || cleanTokenSurface(tok) == "" {
			numrPos = -1
			continue
		}
		cleanLower := strings.ToLower(cleanTokenSurface(tok))

		// noun:numr / number force-noun lookahead
		if HasPosTagRE(tok, numrNounNumrAllRE) {
			if i < len(tokens)-1 && tokens[i+1] != nil &&
				nounForcePattern.MatchString(strings.ToLower(cleanTokenSurface(tokens[i+1]))) {
				numrPos = i
				numrTok = tok
				numrTags = CollectPOSTags(tok)
				isNumber = HasPosTagStart(tok, "number")
				continue
			}
			if i < len(tokens)-2 && tokens[i+1] != nil && tokens[i+2] != nil &&
				HasPosTagRE(tokens[i+1], regexp.MustCompile(`^adj:p:v_rod.*`)) &&
				nounForcePattern.MatchString(strings.ToLower(cleanTokenSurface(tokens[i+2]))) {
				numrPos = i
				numrTok = tok
				numrTags = CollectPOSTags(tok)
				isNumber = HasPosTagStart(tok, "number")
				i++ // skip adj
				continue
			}
		}

		if hasNumrPOS(tok) {
			numrPos = -1
			numrTok = nil
			numrTags = nil
			isNumber = false

			if numrDashLetterRE.MatchString(cleanTokenSurface(tok)) {
				continue
			}
			if HasLemmaToken(tok, "мати") && HasPosTagStart(tok, "verb") {
				continue
			}
			if HasLemmaToken(tok, "один") {
				continue
			}
			for _, p := range CollectPOSTags(tok) {
				if strings.HasPrefix(p, "numr") || numrNounNumrAllRE.MatchString(p) {
					numrPos = i
					numrTok = tok
					numrTags = append(numrTags, p)
				}
			}
			continue
		} else if HasPosTagStart(tok, "number") || isNumberToken(tok) {
			numrPos = i
			numrTok = tok
			numrTags = CollectPOSTags(tok)
			isNumber = true
			continue
		}

		if numrPos < 0 || numrTok == nil {
			continue
		}

		// два з половиною …
		if i < len(tokens)-2 && numrHalfPrepRE.MatchString(cleanLower) &&
			tokens[i+1] != nil && numrHalfNounRE.MatchString(strings.ToLower(cleanTokenSurface(tokens[i+1]))) {
			i++
			continue
		}

		// skip adj for 2-4 + m:v_rod …а/я
		numrClean := strings.ToLower(cleanTokenSurface(numrTok))
		if i < len(tokens)-1 &&
			(matches2to4(numrClean) || numrDva34Pattern.MatchString(numrClean)) &&
			HasPosTagRE(tok, numrAdjPRodNazRE) &&
			HasPosTagAndToken(tokens[i+1], numrMRodAyaRE, numrAyaTokenRE) {
			continue
		}

		// півтора/fract + раз → special message
		if numrPivtoraRE.MatchString(numrClean) || numrFractRE.MatchString(numrClean) {
			if numrRazRE.MatchString(cleanLower) {
				msg := "Після десяткового дробу або «півтора» треба вживати «раза»"
				m := rules.NewRuleMatch(r, sentence, numrTok.GetStartPos(), tok.GetEndPos(), msg)
				m.ShortMessage = r.shortMsg
				out = append(out, m)
				numrPos = -1
				continue
			}
		}

		// «тон» → «тонн»
		if cleanLower == "тон" {
			msg := "Ви мали на увазі: «тонн»?"
			m := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(), msg)
			m.ShortMessage = r.shortMsg
			out = append(out, m)
			numrPos = -1
			continue
		}

		// collect noun/adj readings
		var nounTags []string
		clear := false
		for _, p := range CollectPOSTags(tok) {
			if p == "" || strings.HasSuffix(p, "_END") {
				continue
			}
			if numrNounIgnoreRE.MatchString(p) {
				clear = true
				break
			}
			if strings.HasPrefix(p, "noun") || strings.HasPrefix(p, "adj") {
				nounTags = append(nounTags, p)
			} else if !isPredictOrInsertPOS(p) {
				clear = true
				break
			}
		}

		// багато limited to m:v_rod / force nouns
		if strings.HasSuffix(numrClean, "багато") {
			if !hasMaleUA(tok) && !nounForcePattern.MatchString(cleanLower) {
				numrPos = -1
				continue
			}
		}

		if clear || len(nounTags) == 0 {
			numrPos = -1
			continue
		}

		// build master inflections
		var master []Inflection
		if numrPos == i-2 && i > 0 && tokens[i-1] != nil &&
			numrDesyatykhRE.MatchString(strings.ToLower(cleanTokenSurface(tokens[i-1]))) {
			master = []Inflection{
				{Gender: "m", Case: "v_rod"},
				{Gender: "f", Case: "v_rod"},
				{Gender: "n", Case: "v_rod"},
			}
		} else if isNumber {
			var ok bool
			master, ok = buildNumberMaster(numrClean, tokens, i, cleanLower)
			if !ok {
				numrPos = -1
				continue
			}
		} else {
			if hasNumrPOS(numrTok) {
				master = GetNumrCaseInflections(numrTags)
			} else {
				master = []Inflection{{Gender: "p", Case: "v_rod"}}
			}
			master = adjustAlphaNumrMaster(master, numrClean, tokens, i, nounTags)
		}

		slave := GetNounInflectionsFromTags(nounTags, nil)
		slave = append(slave, GetAdjCaseInflections(nounTags)...)
		// dedupe by String key
		slave = dedupeInflections(slave)

		if !InflectionsIntersect(master, slave) {
			if IsNumrNounException(tokens, numrPos, i) {
				numrPos = -1
				continue
			}
			msg := numrNounMsg(numrTok, numrClean, master, slave, tok, nounTags)
			m := rules.NewRuleMatch(r, sentence, numrTok.GetStartPos(), tok.GetEndPos(), msg)
			m.ShortMessage = r.shortMsg
			out = append(out, m)
		}
		numrPos = -1
	}
	return out
}

func buildNumberMaster(numrClean string, tokens []*languagetool.AnalyzedTokenReadings, i int, cleanLower string) ([]Inflection, bool) {
	// RE2-safe 2_5 / 2to4 (Java lookbehind (?<!1))
	if matches5_5(numrClean) {
		return []Inflection{
			{Gender: "p", Case: "v_rod"},
			{Gender: "m", Case: "v_rod"},
			{Gender: "f", Case: "v_rod"},
			{Gender: "n", Case: "v_rod"},
		}, true
	}
	if matches2_5(numrClean) {
		return []Inflection{
			{Gender: "p", Case: "v_naz"},
			{Gender: "p", Case: "v_zna", AnimTag: "inanim"},
			{Gender: "m", Case: "v_rod"},
			{Gender: "f", Case: "v_rod"},
			{Gender: "n", Case: "v_rod"},
		}, true
	}
	if numr1_5RE.MatchString(numrClean) {
		return []Inflection{
			{Gender: "m", Case: "v_rod"},
			{Gender: "f", Case: "v_rod"},
			{Gender: "n", Case: "v_rod"},
		}, true
	}
	if numrFractRE.MatchString(numrClean) {
		return []Inflection{
			{Gender: "m", Case: "v_rod"},
			{Gender: "f", Case: "v_rod"},
			{Gender: "n", Case: "v_rod"},
		}, true
	}
	if matches2to4(numrClean) && HasPosTagAndToken(tokens[i], numrMRodAyaRE, numrAyaTokenRE) {
		if isNynCase(tokens, i) {
			return []Inflection{{Gender: "m", Case: "v_rod"}}, true
		}
		return []Inflection{
			{Gender: "p", Case: "v_naz"},
			{Gender: "p", Case: "v_zna"},
		}, true
	}
	if matches2to4(numrClean) {
		if isNynCase(tokens, i) {
			return []Inflection{{Gender: "m", Case: "v_rod"}}, true
		}
		return nil, false
	}
	if numr5to9RE.MatchString(numrClean) && nounForcePattern.MatchString(cleanLower) {
		return []Inflection{{Gender: "p", Case: "v_rod"}}, true
	}
	return nil, false
}

func adjustAlphaNumrMaster(master []Inflection, numrToken string, tokens []*languagetool.AnalyzedTokenReadings, i int, nounTags []string) []Inflection {
	// find p:v_naz / p:v_zna
	var pVnazZna []Inflection
	for _, inf := range master {
		if inf.Gender == "p" && (inf.Case == "v_naz" || inf.Case == "v_zna") {
			pVnazZna = append(pVnazZna, inf)
		}
	}
	if len(pVnazZna) == 0 {
		// півтора / півтори without p:v_naz
		if numrPivtoraRE.MatchString(numrToken) {
			return []Inflection{
				{Gender: "m", Case: "v_rod"},
				{Gender: "n", Case: "v_rod"},
			}
		}
		if numrPivtoryRE.MatchString(numrToken) {
			return []Inflection{{Gender: "f", Case: "v_rod"}}
		}
		return master
	}

	removePV := func(src []Inflection) []Inflection {
		var out []Inflection
		for _, inf := range src {
			if inf.Gender == "p" && (inf.Case == "v_naz" || inf.Case == "v_zna") {
				continue
			}
			out = append(out, inf)
		}
		return out
	}

	if numr5to9AlphaRE.MatchString(numrToken) {
		master = removePV(master)
		master = append(master, Inflection{Gender: "p", Case: "v_rod"})
		return master
	}
	if numrDvoeEtcRE.MatchString(numrToken) {
		master = removePV(master)
		master = append(master, Inflection{Gender: "p", Case: "v_rod"})
		return master
	}
	if numrBagatoRE.MatchString(numrToken) {
		master = removePV(master)
		master = append(master,
			Inflection{Gender: "p", Case: "v_rod"},
			Inflection{Gender: "m", Case: "v_rod"},
			Inflection{Gender: "n", Case: "v_rod"},
			Inflection{Gender: "f", Case: "v_rod"},
		)
		return master
	}
	if numrPivRE.MatchString(numrToken) {
		return []Inflection{
			{Gender: "m", Case: "v_rod"},
			{Gender: "f", Case: "v_rod"},
			{Gender: "n", Case: "v_rod"},
		}
	}
	if numrDva34Pattern.MatchString(numrToken) {
		master = removePV(master)
		if isNynCase(tokens, i) {
			master = append(master, Inflection{Gender: "m", Case: "v_rod"})
			if numrToken == "обидва" || numrToken == "обидві" {
				master = append(master, Inflection{Gender: "p", Case: "v_naz"})
			}
		} else {
			master = append(master, Inflection{Gender: "p", Case: "v_naz"})
			if hasTagRE(nounTags, numrInanimPZnaRE) {
				master = append(master, Inflection{Gender: "p", Case: "v_zna"})
			} else if hasTagRE(nounTags, numrAdjPZnaRE) {
				if i == len(tokens)-1 || tokens[i+1] == nil || !HasPosTagRE(tokens[i+1], numrNounPRodRE) {
					master = append(master, Inflection{Gender: "p", Case: "v_zna"})
				}
			}
		}
		return master
	}
	if numrPivtoraRE.MatchString(numrToken) {
		return []Inflection{
			{Gender: "m", Case: "v_rod"},
			{Gender: "n", Case: "v_rod"},
		}
	}
	if numrPivtoryRE.MatchString(numrToken) {
		return []Inflection{{Gender: "f", Case: "v_rod"}}
	}
	return master
}

func numrNounMsg(numrTok *languagetool.AnalyzedTokenReadings, numrClean string, master, slave []Inflection, nounTok *languagetool.AnalyzedTokenReadings, nounTags []string) string {
	numrSurf := ""
	if numrTok != nil && len(numrTok.GetReadings()) > 0 {
		numrSurf = numrTok.GetToken()
	}
	nounSurf := ""
	if nounTok != nil {
		nounSurf = nounTok.GetToken()
	}
	msg := "Потенційна помилка: числівник не узгоджений з іменником: \"" + numrSurf + "\" і \"" + nounSurf + "\""
	if numr1_5RE.MatchString(numrClean) {
		return "Після «1,5» треба вживати родовий відмінок однини"
	}
	if matches2_5(numrClean) {
		return "Після числівника, що закінчується на 2-4 і потім «,5», іменник має стояти в називному відмінку множини (якщо вимовляємо «з половиною»), або в родовом відмінку однини (якщо вимовляємо «і п'ять десятих»)"
	}
	if strings.HasSuffix(numrClean, ",5") {
		return "Після числівника, що закінчується на 5-9 і потім «,5», іменник має стояти в родовому відмінку множини (якщо вимовляємо «з половиною»), або в родовом відмінку однини (якщо вимовляємо «і п'ять десятих»)"
	}
	if strings.EqualFold(numrClean, "півтора") {
		return "Існує правило, що після «півтора» треба вживати родовий відмінок ч. або с.р., однак у текстах в багатьох випадках вживають і форму множини, надто коли перед іменником іде прикметник"
	}
	if strings.EqualFold(numrClean, "півтори") {
		return "Існує правило, що після «півтора» треба вживати родовий відмінок ж.р., однак у текстах в багатьох випадках вживають і форму множини, надто коли перед іменником іде прикметник"
	}
	if hasInflection(master, "m", "v_rod") && numrUYuyuLike(nounTok) && hasTagRE(nounTags, numrDavMRE) {
		msg += UsedUInsteadOfAMsg
	}
	return msg
}

func numrUYuyuLike(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	return regexp.MustCompile(`.*[ую]$`).MatchString(tok.GetToken())
}

func hasInflection(infs []Inflection, gender, cas string) bool {
	for _, inf := range infs {
		if inf.Gender == gender && inf.Case == cas {
			return true
		}
	}
	return false
}

func hasTagRE(tags []string, re *regexp.Regexp) bool {
	for _, p := range tags {
		if re.MatchString(p) {
			return true
		}
	}
	return false
}

func dedupeInflections(infs []Inflection) []Inflection {
	seen := map[string]struct{}{}
	var out []Inflection
	for _, inf := range infs {
		k := inf.String()
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, inf)
	}
	return out
}

// HasPosTagAndToken ports PosTagHelper.hasPosTagAndToken.
func HasPosTagAndToken(tok *languagetool.AnalyzedTokenReadings, posRE, tokenRE *regexp.Regexp) bool {
	if tok == nil || posRE == nil || tokenRE == nil {
		return false
	}
	surf := cleanTokenSurface(tok)
	if !tokenRE.MatchString(surf) {
		return false
	}
	return HasPosTagRE(tok, posRE)
}

// isNynCase ports TokenAgreementNumrNounRule.isNynCase (lemma/surface -нин(а/я)).
func isNynCase(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if i < 0 || i >= len(tokens) || tokens[i] == nil {
		return false
	}
	tok := tokens[i]
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil || r.GetLemma() == nil {
			continue
		}
		pos, lem := *r.GetPOSTag(), *r.GetLemma()
		surf := strings.ToLower(r.GetToken())
		if numrNynRodRE.MatchString(pos) && numrNynRodTokenRE.MatchString(surf) {
			base := regexp.MustCompile(`[ая]$`).ReplaceAllString(surf, "")
			if lem == base {
				return true
			}
		}
		if numrNynNazRE.MatchString(pos) && numrNynNazTokenRE.MatchString(surf) {
			base := regexp.MustCompile(`ни$`).ReplaceAllString(surf, "нин")
			if lem == base {
				return true
			}
		}
	}
	return false
}

// hasMaleUA ports PosTagHelper.hasMaleUA (m + UA noun soft path for багато).
func hasMaleUA(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	for _, p := range CollectPOSTags(tok) {
		if strings.HasPrefix(p, "noun") && strings.Contains(p, ":m:") &&
			(strings.Contains(p, "v_rod") || strings.Contains(p, "v_naz")) {
			return true
		}
	}
	return false
}

// RE2-safe numeric helpers (Java (?<!1) lookbehind).
func matches2to4(s string) bool {
	// ([0-9]+[–-])?[^,]*(?<!1)[234]
	if s == "" {
		return false
	}
	// strip optional leading N–
	s2 := s
	if i := strings.LastIndexAny(s, "–-"); i >= 0 && i < len(s)-1 {
		// only strip range prefix if left part is digits
		left := s[:i]
		if left != "" && isAllDigits(left) {
			s2 = s[i+1:]
		}
	}
	if strings.Contains(s2, ",") {
		return false
	}
	if len(s2) == 0 {
		return false
	}
	last := s2[len(s2)-1]
	if last != '2' && last != '3' && last != '4' {
		return false
	}
	// (?<!1) — not preceded by 1
	if len(s2) >= 2 && s2[len(s2)-2] == '1' {
		return false
	}
	return true
}

func matches2_5(s string) bool {
	// .*(?<!1)[234],5
	if !strings.HasSuffix(s, ",5") {
		return false
	}
	base := strings.TrimSuffix(s, ",5")
	if base == "" {
		return false
	}
	last := base[len(base)-1]
	if last != '2' && last != '3' && last != '4' {
		return false
	}
	if len(base) >= 2 && base[len(base)-2] == '1' {
		return false
	}
	return true
}

func matches5_5(s string) bool {
	return numr5_5RE.MatchString(s)
}

func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return s != ""
}
