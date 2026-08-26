package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// TokenAgreementNumrNounExceptionHelper ports
// org.languagetool.rules.uk.TokenAgreementNumrNounExceptionHelper.
type TokenAgreementNumrNounExceptionHelper struct{}

func NewTokenAgreementNumrNounExceptionHelper() *TokenAgreementNumrNounExceptionHelper {
	return &TokenAgreementNumrNounExceptionHelper{}
}

// IsException ports isException(tokens, state, numrInflections, slaveInflections, slaveTokenReadings).
// master is State-derived masterInflections (Java param name numrInflections).
// isNumber is State.number. slaveInflections / slaveTokenReadings are unused in Java body.
func (TokenAgreementNumrNounExceptionHelper) IsException(
	tokens []*languagetool.AnalyzedTokenReadings,
	numrPos, nounPos int,
	master []Inflection,
	isNumber bool,
) bool {
	return IsNumrNounException(tokens, numrPos, nounPos, master, isNumber)
}

// Exception is a thin wrapper keeping the historical two-index call shape (derives master from tags).
func (TokenAgreementNumrNounExceptionHelper) Exception(tokens []*languagetool.AnalyzedTokenReadings, numrPos, nounPos int) bool {
	return IsNumrNounException(tokens, numrPos, nounPos, nil, false)
}

// Patterns from TokenAgreementNumrNounExceptionHelper (Java String.matches / Pattern full-string).
var (
	numrManyObPattern = regexp.MustCompile(
		`^(?:багать(?:ох|ом|ма)|обо(?:х|м|ма)|(?:дв|трь|чотирь)о[хм]|скільки(?:сь)?(?:-небудь)?|стільки)$`)
	numrNounSoftSurface = regexp.MustCompile(
		`^(?:плюс|мінус|ранку|вечора|ночі|тепла|морозу|родом|зростом|дивом|станом|вагою|слід|типу|формату|вартістю|році|населення)$`)
	numrLemmaSoftRE = regexp.MustCompile(
		`^(?:у?весь|який(?:сь)?|свій|сам|цей|решта|кількість|вартий|кожний|жодний|менший|більший|вищий|нижчий)$`)
	numrPivtoraFullRE = regexp.MustCompile(`^(?:(?:один-|одне-)?півтора|(?:одна-)?півтори)$`)
	numrArticlePrevRE = regexp.MustCompile(
		`^(?:ч\.|ст\.|п\.|частина|стаття|пункт|підпункт|абзац|№|номер)$`)
	numrFractHalfPrepRE = regexp.MustCompile(
		`^(?:від|до|протягом|[ув]продовж|близько|після|для|більше|менше)$`)
	numrObyeLikeRE = regexp.MustCompile(`^(?:обоє|двоє|троє|.+еро)$`)
	numrObyeAnimRE = regexp.MustCompile(`^(?:обоє|обидвоє|троє)$`)
	numrSyomaRE    = regexp.MustCompile(`^(?:сьома|дев.яноста)$`)
	numrNextSoftRE = regexp.MustCompile(`^(?:[.,:;()«»—–-]|і|й|та)$`)
)

// IsNumrNounException ports TokenAgreementNumrNounExceptionHelper.isException.
// master is the rule's masterInflections (Java numrInflections parameter).
// When master is nil, numr POS tags are used (for surface-only call sites).
// isNumber is State.number.
func IsNumrNounException(
	tokens []*languagetool.AnalyzedTokenReadings,
	numrPos, nounPos int,
	master []Inflection,
	isNumber bool,
) bool {
	// Invalid layout → exception (no flag). Mirrors defensive ports of sibling helpers.
	if numrPos < 0 || nounPos <= numrPos || numrPos >= len(tokens) || nounPos >= len(tokens) {
		return true
	}
	numr, noun := tokens[numrPos], tokens[nounPos]
	if numr == nil || noun == nil {
		return true
	}
	numrLower := CleanTokenLower(numr)
	nounLower := CleanTokenLower(noun)

	// Java: numrInflections = masterInflections from the rule
	numrInfs := master
	if numrInfs == nil {
		numrInfs = GetNumrCaseInflections(CollectPOSTags(numr))
	}

	// для багатьох несподіванкою стало
	// numrTokenLower.matches("багать(ох|ом|ма)|обо(х|м|ма)|(дв|трь|чотирь)о[хм]|скільки(сь)?(-небудь)?|стільки")
	if numrManyObPattern.MatchString(numrLower) {
		return true
	}

	// |десятих|сотих|тисячних - done in main rule
	// nounLowerToken.matches("плюс|мінус|ранку|…")
	if numrNounSoftSurface.MatchString(nounLower) {
		return true
	}

	// рази може у два більший
	// LemmaHelper.hasLemma(noun, Pattern.compile("у?весь|який(сь)?|…"))
	if HasLemmaTokenRE(noun, numrLemmaSoftRE) {
		return true
	}

	// хвилин п'ять люди / сотні дві персон
	// TIME_PLUS before numr with non-disjoint master vs prev noun inflections
	if numrPos > 1 &&
		HasLemmaWithPosRE(tokens[numrPos-1], TimePlusLemmaList(), regexp.MustCompile(`noun.*?.:v_(naz|rod).*`)) {
		prevInfs := GetNounCaseInflections(CollectPOSTags(tokens[numrPos-1]))
		if InflectionsIntersect(numrInfs, prevInfs) {
			return true
		}
	}

	// півтора довгих роки
	if nounPos < len(tokens)-1 &&
		numrPivtoraFullRE.MatchString(CleanTokenLower(numr)) &&
		HasPosTagRE(noun, regexp.MustCompile(`adj:p:v_(naz|rod).*`)) &&
		HasPosTagRE(tokens[nounPos+1], regexp.MustCompile(`noun.*?:p:v_naz.*`)) {
		return true
	}

	// хвилин зо п'ять люди — TIME_PLUS + prep + numr
	if numrPos > 2 &&
		HasPosTagStart(tokens[numrPos-1], "prep") &&
		HasLemmaWithPosRE(tokens[numrPos-2], TimePlusLemmaList(), regexp.MustCompile(`noun.*?p:v_(naz|rod).*`)) {
		prevInfs := GetNounCaseInflections(CollectPOSTags(tokens[numrPos-2]))
		if InflectionsIntersect(numrInfs, prevInfs) {
			return true
		}
	}

	// У свої вісімдесят пан Василь
	if numrPos > 2 &&
		HasPosTagStart(tokens[numrPos-2], "prep") &&
		CleanTokenLower(tokens[numrPos-1]) == "свої" &&
		HasPosTagRE(numr, regexp.MustCompile(`numr:p:v_zna.*`)) &&
		HasPosTagRE(noun, regexp.MustCompile(`noun:anim:.:v_naz.*`)) {
		return true
	}

	// два провінційного вигляду персонажі
	// Java: noun:inanim:.:v_rod(?!.*pron) / noun(?!.*pron) — RE2: filter :pron
	if nounPos <= len(tokens)-3 &&
		HasPosTagRE(noun, regexp.MustCompile(`adj:.:v_rod.*`)) &&
		hasPosWithoutPron(tokens[nounPos+1], regexp.MustCompile(`noun:inanim:.:v_rod`)) &&
		hasPosWithoutPron(tokens[nounPos+2], regexp.MustCompile(`^noun`)) {
		adjG := gendersFromPos(noun, regexp.MustCompile(`adj:([mfnp]):v_rod`))
		nounG := gendersFromPos(tokens[nounPos+1], regexp.MustCompile(`noun:inanim:([mfnp]):v_rod`))
		if adjG != "" && nounG != "" && gendersOverlap(adjG, nounG) {
			realInfs := GetNounCaseInflections(CollectPOSTags(tokens[nounPos+2]))
			if InflectionsIntersect(numrInfs, realInfs) {
				return true
			}
		}
	}

	// handled by another rule: ,5 + тон|тис|коп or left prep-ish
	if strings.HasSuffix(numrLower, ",5") {
		if regexp.MustCompile(`^(?:тон|тис|коп)$`).MatchString(nounLower) {
			return true
		}
		if numrPos > 1 && numrFractHalfPrepRE.MatchString(CleanTokenLower(tokens[numrPos-1])) {
			return true
		}
	}

	// обоє горбаті — adj:p:v_naz + surface .+і
	// Java: hasPosTagAndToken(…, adj:p:v_naz.*, ".+і") — reading token matches
	if numrObyeLikeRE.MatchString(numrLower) {
		if hasPosTagAndTokenReading(noun, regexp.MustCompile(`adj:p:v_naz.*`), regexp.MustCompile(`.+і`)) {
			return true
		}
	}

	// обоє режисери
	if numrObyeAnimRE.MatchString(numrLower) &&
		HasPosTagRE(noun, regexp.MustCompile(`noun:anim:p:v_naz.*`)) {
		return true
	}

	// 22 червня — state.number + MONTH :m:v_rod
	if isNumber && HasLemmaWithPartPos(noun, MonthLemmas, ":m:v_rod") {
		return true
	}

	// 3 / 4 понеділка
	if numrPos > 2 && tokens[numrPos-1] != nil && tokens[numrPos-1].GetCleanToken() == "/" {
		return true
	}

	// ч.|ст.|…|№ before numr
	// Java: hasLemma(…, Arrays.asList("ч.", "ст.", …)) || token equals "№"
	if numrPos > 1 {
		prev := tokens[numrPos-1]
		if prev != nil {
			if HasLemmaTokenAny(prev, []string{"ч.", "ст.", "п.", "частина", "стаття", "пункт", "підпункт", "абзац", "№", "номер"}) ||
				prev.GetCleanToken() == "№" ||
				numrArticlePrevRE.MatchString(CleanTokenLower(prev)) {
				return true
			}
		}
	}

	// двадцять перший; дві соті
	if HasPosTagRE(noun, regexp.MustCompile(`adj.*numr.*`)) {
		return true
	}

	// два нових горнятка / 2 хворих/злотих
	if numrDva34Pattern.MatchString(numrLower) || isNumber {
		// adj(?!.*numr).*:p:v_rod.*
		if hasAdjPRodNotNumr(noun) {
			if nounPos == len(tokens)-1 {
				return true
			}
			next := tokens[nounPos+1]
			if next != nil {
				// adj(?!.*numr).*:p:v_rod.*|noun.*:p:v_naz.*|prep
				// || ! (adj|noun).* || token matches punctuation/conj
				if hasAdjPRodNotNumr(next) ||
					HasPosTagRE(next, regexp.MustCompile(`noun.*:p:v_naz.*`)) ||
					HasPosTagStart(next, "prep") ||
					!HasPosTagRE(next, regexp.MustCompile(`(?:adj|noun).*`)) ||
					numrNextSoftRE.MatchString(next.GetCleanToken()) {
					return true
				}
			}
		}
		// 2 хворих/злотих...
		if strings.HasSuffix(nounLower, "их") &&
			HasPosTagRE(noun, regexp.MustCompile(`noun.*:p:v_rod.*`)) {
			return true
		}
	}

	// сьома вода
	if numrSyomaRE.MatchString(numrLower) &&
		HasPosTagRE(noun, regexp.MustCompile(`(?:noun:.*?|adj):[fp]:v_naz.*`)) {
		return true
	}

	return false
}

// hasPosTagAndTokenReading ports PosTagHelper.hasPosTagAndToken (per-reading surface).
func hasPosTagAndTokenReading(tok *languagetool.AnalyzedTokenReadings, posRE, tokenRE *regexp.Regexp) bool {
	if tok == nil || posRE == nil || tokenRE == nil {
		return false
	}
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		if posRE.MatchString(*r.GetPOSTag()) && tokenRE.MatchString(r.GetToken()) {
			return true
		}
	}
	return false
}
