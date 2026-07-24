package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// ConjForPluralWithComma ports TokenAgreementAdjNounExceptionHelper.CONJ_FOR_PLURAL_WITH_COMMA
// (incl. Latin a/i and comma).
var ConjForPluralWithComma = []string{
	"і", "а", "й", "та", "чи", "або", "ані", "також", "плюс", "то", "a", "i", ",",
}

// ConjForPlural ports CONJ_FOR_PLURAL (no comma/плюс).
var ConjForPlural = []string{
	"і", "а", "й", "та", "чи", "або", "ані", "також", "то", "a", "i",
}

// DovyeTroyeRE ports DOVYE_TROYE (two/three/four numeral surfaces; full-string).
var DovyeTroyeRE = regexp.MustCompile(
	`^(?:.*[2-4]|.*[2-4][\x{2013}\x{2014}-].*[2-4]|два|обидва|двоє|двійко|три|троє|чотири|один[\x{2013}\x{2014}-]два|два[\x{2013}\x{2014}-]три|три[\x{2013}\x{2014}-]чотири|двоє[\x{2013}\x{2014}-]троє|троє[\x{2013}\x{2014}-]четверо|півтор[аи])$`)

// IsNonPluralA ports TokenAgreementNounVerbExceptionHelper.isNonPluralA.
func IsNonPluralA(tokens []*languagetool.AnalyzedTokenReadings, pos int) bool {
	if pos < 0 || pos >= len(tokens) || tokens[pos] == nil {
		return false
	}
	t := tokens[pos].GetToken()
	if t != "а" && t != "a" {
		return false
	}
	if pos+1 >= len(tokens) {
		return true
	}
	return !HasLemmaTokenAny(tokens[pos+1], []string{"також", "потім", "пізніше"})
}

func isConjForPluralWithComma(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	return containsStr(ConjForPluralWithComma, strings.ToLower(tok.GetToken()))
}

// reverseConjAdvFind ports TokenAgreementAdjNounExceptionHelper.reverseConjAdvFind.
func reverseConjAdvFind(tokens []*languagetool.AnalyzedTokenReadings, pos, depth int) bool {
	for i := pos; i > pos-depth && i >= 2; i-- {
		if tokens[i] == nil {
			continue
		}
		if isConjForPluralWithComma(tokens[i]) {
			// Java: adv(?!p) left or (adv(?!p)|part) right
			if hasAdvNotAdvp(tokens[i-1]) || hasAdvNotAdvp(tokens[i+1]) || HasPosTagStart(tokens[i+1], "part") {
				return true
			}
		}
		if HasPosTagPart(tokens[i], "verb") {
			return false
		}
	}
	return false
}

// reverseConjFind ports reverseConjFind.
func reverseConjFind(tokens []*languagetool.AnalyzedTokenReadings, pos, depth int) bool {
	for i := pos; i > pos-depth && i >= 1; i-- {
		if tokens[i] == nil {
			continue
		}
		if isConjForPluralWithComma(tokens[i]) {
			if i < 2 || !HasPosTagRE(tokens[i-1], regexp.MustCompile(`^(?:adj|numr|conj:coord)`)) {
				return false
			}
			return true
		}
		if i >= 1 && tokens[i-1] != nil &&
			!hasAdjConjNumPrepAdv(tokens[i-1]) &&
			tokens[i-1].GetToken() != "," {
			return false
		}
	}
	return false
}

// reverseConjFind2 ports reverseConjFind2 (looser left-context for number lists).
func reverseConjFind2(tokens []*languagetool.AnalyzedTokenReadings, pos, depth int) bool {
	for i := pos; i > pos-depth && i >= 1; i-- {
		if tokens[i] == nil {
			continue
		}
		if isConjForPluralWithComma(tokens[i]) {
			if IsNonPluralA(tokens, i) {
				return false
			}
			if i < 2 {
				return false
			}
			left := tokens[i-1]
			right := (*languagetool.AnalyzedTokenReadings)(nil)
			if i+1 < len(tokens) {
				right = tokens[i+1]
			}
			// Java returns false when left is NOT one of the special coordination markers.
			// Specials: (number + adj numr right) OR comma OR dash OR paren OR /або OR unknown.
			special := false
			if left != nil {
				if left.HasPosTag("number") && right != nil && HasPosTagRE(right, regexp.MustCompile(`adj.*?numr.*`)) {
					special = true
				}
				if left.GetToken() == "," {
					special = true
				}
				if regexp.MustCompile(`.*[–-]`).MatchString(left.GetToken()) {
					special = true
				}
				if regexp.MustCompile(`^[)»”]$`).MatchString(left.GetToken()) {
					special = true
				}
				if left.GetToken() == "/" && tokens[i].GetToken() == "або" {
					special = true
				}
				if isUnknownWord(left) {
					special = true
				}
			}
			if !special {
				return false
			}
			return true
		}
		if i >= 1 && tokens[i-1] != nil &&
			!hasAdjConjNumPrepAdv(tokens[i-1]) &&
			tokens[i-1].GetToken() != "," {
			return false
		}
	}
	return false
}

func hasAdjConjNumPrepAdv(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	if HasPosTagRE(tok, regexp.MustCompile(`^(?:adj|conj:coord|num|prep)`)) {
		return true
	}
	return hasAdvNotAdvp(tok)
}

func isUnknownWord(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return true
	}
	tags := CollectPOSTags(tok)
	if len(tags) == 0 {
		return true
	}
	for _, p := range tags {
		if p == "" || strings.HasPrefix(p, "unknown") || p == "_" {
			continue
		}
		if strings.Contains(p, "SENT") {
			continue
		}
		return false
	}
	return true
}

// forwardConjFind ports forwardConjFind.
func forwardConjFind(tokens []*languagetool.AnalyzedTokenReadings, pos, depth int) bool {
	for i := pos; i < len(tokens) && i <= pos+depth; i++ {
		if tokens[i] == nil {
			continue
		}
		if isConjForPluralWithComma(tokens[i]) {
			if i < len(tokens)-3 && checkTextInSent(tokens, i+1, "а також") {
				if HasPosTagRE(tokens[i+3], regexp.MustCompile(`^(?:noun|adj|num)`)) || hasAdvNotAdvp(tokens[i+3]) {
					return true
				}
			}
			if i == len(tokens)-1 {
				return false
			}
			next := tokens[i+1]
			if next == nil {
				return false
			}
			if !HasPosTagRE(next, regexp.MustCompile(`^(?:noun|adj|num)`)) &&
				!hasAdvNotAdvp(next) &&
				!IsCapitalized(next.GetCleanToken()) &&
				!regexp.MustCompile(`^["«“„]`).MatchString(next.GetToken()) {
				return false
			}
			return true
		}
		if !HasPosTagRE(tokens[i], regexp.MustCompile(`^(?:noun|adj|prep|number:latin)`)) &&
			!hasAdvNotAdvp(tokens[i]) &&
			!IsCapitalized(tokens[i].GetCleanToken()) {
			return false
		}
	}
	return false
}

func checkTextInSent(tokens []*languagetool.AnalyzedTokenReadings, pos int, text string) bool {
	// Java: text.split(" "); for (i < words.length && i+pos < tokens.length)
	// returns true if every compared pair matches (including when words outrun tokens).
	words := strings.Split(text, " ")
	for i := 0; i < len(words) && i+pos < len(tokens); i++ {
		if tokens[i+pos] == nil || !strings.EqualFold(tokens[i+pos].GetToken(), words[i]) {
			return false
		}
	}
	return true
}

// TokenAgreementAdjNounExceptionHelper anchors the Java twin name; logic is IsAdjNounException.
type TokenAgreementAdjNounExceptionHelper struct{}

func (TokenAgreementAdjNounExceptionHelper) IsException(tokens []*languagetool.AnalyzedTokenReadings, adjPos, nounPos int) bool {
	return IsAdjNounException(tokens, adjPos, nounPos)
}
