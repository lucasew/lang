package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// MixedAlphabetsRule ports org.languagetool.rules.uk.MixedAlphabetsRule.
// A rule that matches words with Latin and Cyrillic characters in them.
//
// Java: setCategory(Categories.MISC.getCategory(messages)).
type MixedAlphabetsRule struct {
	Messages map[string]string
	Category *rules.Category
}

// NewMixedAlphabetsRule ports MixedAlphabetsRule(ResourceBundle messages).
func NewMixedAlphabetsRule(messages map[string]string) *MixedAlphabetsRule {
	return &MixedAlphabetsRule{
		Messages: messages,
		Category: rules.CatMisc.GetCategory(messages),
	}
}

// GetID ports MixedAlphabetsRule.getId → UK_MIXED_ALPHABETS.
func (r *MixedAlphabetsRule) GetID() string { return "UK_MIXED_ALPHABETS" }

// GetDescription ports MixedAlphabetsRule.getDescription.
func (r *MixedAlphabetsRule) GetDescription() string {
	return "Змішування кирилиці й латиниці"
}

// getShort ports MixedAlphabetsRule.getShort.
func (r *MixedAlphabetsRule) getShort() string {
	return "Мішанина розкладок"
}

// GetCategory ports Rule.getCategory (Java MISC).
func (r *MixedAlphabetsRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// Pattern constants — Java Pattern.matcher(...).matches() = full string (anchored in Go).
var (
	// LIKELY_LATIN_NUMBER = [XVIХІ]{2,8}(-[а-яіїє]{1,3})?
	likelyLatinNumber = regexp.MustCompile(`^[XVIХІ]{2,8}(?:-[а-яіїє]{1,3})?$`)
	// LATIN_NUMBER_WITH_CYRILLICS = (Х{1,3}І{1,3}|І{1,3}Х{1,3}|Х{2,3}|І{2,3})(-[а-яіїє]{1,4})?
	latinNumberWithCyr = regexp.MustCompile(`^(?:Х{1,3}І{1,3}|І{1,3}Х{1,3}|Х{2,3}|І{2,3})(?:-[а-яіїє]{1,4})?$`)
	// MIXED_ALPHABETS
	mixedAlphabets = regexp.MustCompile(`^.*(?:[a-zA-ZïáÁéÉíÍḯḮóÓúýÝ]'?[а-яіїєґА-ЯІЇЄҐ]|[а-яіїєґА-ЯІЇЄҐ]'?[a-zA-ZïáÁéÉíÍḯḮóÓúýÝ]).*$`)
	// CYRILLIC_ONLY
	cyrillicOnly = regexp.MustCompile(`^.*[бвгґдєжзийїлнпфцчшщьюяБГҐДЄЖЗИЙЇЛПФЦЧШЩЬЮЯ].*$`)
	// LATIN_ONLY
	latinOnly = regexp.MustCompile(`^.*[bdfghjlqrstvzDFGJLNQRSUVZ].*$`)
	// COMMON_CYR_LETTERS = [АВЕІКОРСТУХ]+
	commonCyrLetters = regexp.MustCompile(`^[АВЕІКОРСТУХ]+$`)
	// CYRILLIC_FIRST_LETTER = [а-яіїєґА-ЯІЇЄҐ].*
	cyrillicFirstLetter = regexp.MustCompile(`^[а-яіїєґА-ЯІЇЄҐ].*$`)
	// tokenString.matches("[iya]")
	singleLatinIYA = regexp.MustCompile(`^[iya]$`)
	// token.matches("[xbB]")
	formulaXB = regexp.MustCompile(`^[xbB]$`)
	// prevLemma.matches("гепатит|група|турнір")
	mixedGroupLemmaRE = regexp.MustCompile(`^(?:гепатит|група|турнір)$`)
	// adjustForInvalidSuffix replaceFirst("-[а-яіїє]{1,4}", "")
	invalidSuffixRE = regexp.MustCompile(`-[а-яіїє]{1,4}`)
	// adjustForInvalidSuffix msg: matches("[IVXІХ]+-[а-яіїє]{1,4}")
	invalidSuffixMsgRE = regexp.MustCompile(`^[IVXІХ]+-[а-яіїє]{1,4}$`)
)

// Match ports MixedAlphabetsRule.match.
func (r *MixedAlphabetsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var ruleMatches []*rules.RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()

	for i := 1; i < len(tokens); i++ {
		tokenReadings := tokens[i]
		// Java: tokenReadings.getCleanToken()
		tokenString := tokenReadings.GetCleanToken()

		// optimization: 1-letter tokens first
		// Java: i<tokens.length-1 && (matches [iya] || (A && i==1)) && CYRILLIC_FIRST_LETTER next
		//       && noneMatch token.matches("[xbB]")
		if i < len(tokens)-1 &&
			(singleLatinIYA.MatchString(tokenString) || (tokenString == "A" && i == 1)) &&
			cyrillicFirstLetter.MatchString(tokens[i+1].GetToken()) &&
			!hasFormulaXBToken(tokens) {
			msg := "Вжито латинську «" + tokenString + "» замість кириличної"
			ruleMatches = append(ruleMatches, r.createRuleMatch(tokenReadings, []string{toCyrillic(tokenString)}, msg, sentence))
		} else if tokenString == "І" && likelyBadLatinI(tokens, i) {
			msg := "Вжито кириличну літеру замість латинської"
			ruleMatches = append(ruleMatches, r.createRuleMatch(tokenReadings, []string{toLatin(tokenString)}, msg, sentence))
		} else if i <= len(tokens)-1 && // always true; keep Java guard shape
			tokenString == "І." &&
			(i > 1 &&
				tokens[i-1].GetCleanToken() != "Тому" &&
				tokens[i-1].GetCleanToken() != "Франко" &&
				// PosTagHelper.hasPosTag(prev, "(?!.*:abbr).*fname.*")
				hasFnameNotAbbrPOS(tokens[i-1])) {
			msg := "Вжито кириличну літеру замість латинської"
			ruleMatches = append(ruleMatches, r.createRuleMatch(tokenReadings, []string{toLatin(tokenString)}, msg, sentence))
		} else if commonCyrLetters.MatchString(tokenString) {
			// Java: tokens[i-1].getAnalyzedToken(0).getLemma()
			prevLemma := firstReadingLemma(tokens[i-1])
			if prevLemma != "" && mixedGroupLemmaRE.MatchString(prevLemma) {
				msg := "Вжито кириличну літеру замість латинської"
				ruleMatches = append(ruleMatches, r.createRuleMatch(tokenReadings, []string{toLatin(tokenString)}, msg, sentence))
			}
		}

		// Java: tokenString.length() < 2 (UTF-16 code units)
		if tokenizers.UTF16Len(tokenString) < 2 {
			if tokenString == "°" &&
				i < len(tokens)-1 &&
				tokens[i+1].GetCleanToken() == "С" { // Cyrillic С
				msg := "Вжито кириличну літеру замість латинської"
				ruleMatches = append(ruleMatches, r.createRuleMatch(tokens[i+1], []string{"C"}, msg, sentence))
			}
			continue
		}

		if mixedAlphabets.MatchString(tokenString) {
			msg := "Вжито кириличні й латинські літери в одному слові"
			var replacements []string

			if !latinOnly.MatchString(tokenString) && !likelyLatinNumber.MatchString(tokenString) {
				replacements = append(replacements, toCyrillic(tokenString))
			}
			if (tokenizers.UTF16Len(tokenString) > 2 && !cyrillicOnly.MatchString(tokenString)) ||
				likelyLatinNumber.MatchString(tokenString) {
				converted := toLatinLeftOnly(tokenString)
				converted = adjustForInvalidSuffix(converted)
				replacements = append(replacements, converted)
				msg = "Вжито кириличні літери замість латинських"
				msg = adjustForInvalidSuffixMsg(tokenString, msg)
			}

			if len(replacements) > 0 {
				ruleMatches = append(ruleMatches, r.createRuleMatch(tokenReadings, replacements, msg, sentence))
			}
		} else if latinNumberWithCyr.MatchString(tokenString) {
			converted := toLatinLeftOnly(tokenString)
			converted = adjustForInvalidSuffix(converted)
			msg := "Вжито кириличні літери замість латинських на позначення римської цифри"
			msg = adjustForInvalidSuffixMsg(tokenString, msg)
			ruleMatches = append(ruleMatches, r.createRuleMatch(tokenReadings, []string{converted}, msg, sentence))
		}

		// Java: indexOf('\u0306') > 0 || indexOf('\u0308') > 0
		// (indexOf returns first index; > 0 means not at position 0 and present)
		if idxCombiningAbove(tokenString, '\u0306') || idxCombiningAbove(tokenString, '\u0308') {
			// Java: tokenString.matches(".*(и\u0306|і\u0308).*")
			if strings.Contains(tokenString, "и\u0306") || strings.Contains(tokenString, "і\u0308") {
				fix := strings.ReplaceAll(tokenString, "и\u0306", "й")
				fix = strings.ReplaceAll(fix, "і\u0308", "ї")
				msg := "Вжито комбіновані символи замість українських літер"
				ruleMatches = append(ruleMatches, r.createRuleMatch(tokenReadings, []string{fix}, msg, sentence))
			}
		}
	}

	return ruleMatches
}

// likelyBadLatinI ports MixedAlphabetsRule.likelyBadLatinI.
//
//	i > 1 && (
//	  (isCapitalized(prev) || (hasPosTagStart(prev,"prep") && next not all-upper))
//	  || next in {"ст.","тис."}
//	  || next in {"квартал","півріччя","тисячоліття","половина"}
//	)
func likelyBadLatinI(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if i <= 1 {
		return false
	}
	prevClean := tokens[i-1].GetCleanToken()
	capOrPrep := IsCapitalized(prevClean) ||
		(HasPosTagStart(tokens[i-1], "prep") &&
			i < len(tokens)-1 && !IsAllUppercaseUk(tokens[i+1].GetCleanToken()))
	nextOK := false
	if i < len(tokens)-1 {
		next := tokens[i+1].GetCleanToken()
		if next == "ст." || next == "тис." {
			nextOK = true
		}
		switch next {
		case "квартал", "півріччя", "тисячоліття", "половина":
			nextOK = true
		}
	}
	return capOrPrep || nextOK
}

// adjustForInvalidSuffix ports MixedAlphabetsRule.adjustForInvalidSuffix(String).
// Java: replaceFirst("-[а-яіїє]{1,4}", "") when contains "-".
func adjustForInvalidSuffix(tokenString string) string {
	if strings.Contains(tokenString, "-") {
		loc := invalidSuffixRE.FindStringIndex(tokenString)
		if loc != nil {
			tokenString = tokenString[:loc[0]] + tokenString[loc[1]:]
		}
	}
	return tokenString
}

// adjustForInvalidSuffixMsg ports MixedAlphabetsRule.adjustForInvalidSuffix(String, String).
func adjustForInvalidSuffixMsg(tokenString, msg string) string {
	if strings.Contains(tokenString, "-") && invalidSuffixMsgRE.MatchString(tokenString) {
		msg += ". Також: до римських цифр букви не дописуються."
	}
	return msg
}

// toLatinLeftOnly ports MixedAlphabetsRule.toLatinLeftOnly.
func toLatinLeftOnly(tokenString string) string {
	parts := strings.SplitN(tokenString, "-", 2)
	right := ""
	if len(parts) > 1 {
		right = "-" + parts[1]
	}
	return toLatin(parts[0]) + right
}

// createRuleMatch ports MixedAlphabetsRule.createRuleMatch.
func (r *MixedAlphabetsRule) createRuleMatch(readings *languagetool.AnalyzedTokenReadings, replacements []string, msg string, sentence *languagetool.AnalyzedSentence) *rules.RuleMatch {
	rm := rules.NewRuleMatch(r, sentence, readings.GetStartPos(), readings.GetEndPos(), msg)
	rm.SetShortMessage(r.getShort())
	rm.SetSuggestedReplacements(replacements)
	return rm
}

// firstReadingLemma ports tokens[i-1].getAnalyzedToken(0).getLemma().
func firstReadingLemma(prev *languagetool.AnalyzedTokenReadings) string {
	if prev == nil {
		return ""
	}
	at := prev.GetAnalyzedToken(0)
	if at == nil || at.GetLemma() == nil {
		return ""
	}
	return *at.GetLemma()
}

// hasFnameNotAbbrPOS ports PosTagHelper.hasPosTag(…, Pattern.compile("(?!.*:abbr).*fname.*"))
// Matcher.matches(): full POS contains "fname" and does not contain ":abbr".
// (RE2 has no lookaround; equivalent rewrite of the Java pattern.)
func hasFnameNotAbbrPOS(atr *languagetool.AnalyzedTokenReadings) bool {
	if atr == nil {
		return false
	}
	for _, at := range atr.GetReadings() {
		if at == nil || at.GetPOSTag() == nil {
			continue
		}
		pos := *at.GetPOSTag()
		// full-string matches of (?!.*:abbr).*fname.*
		if strings.Contains(pos, ":abbr") {
			continue
		}
		if strings.Contains(pos, "fname") {
			return true
		}
	}
	return false
}

// hasFormulaXBToken ports Arrays.stream(tokens).noneMatch(t -> t.getToken().matches("[xbB]")) inverted.
func hasFormulaXBToken(tokens []*languagetool.AnalyzedTokenReadings) bool {
	for _, t := range tokens {
		if t != nil && formulaXB.MatchString(t.GetToken()) {
			return true
		}
	}
	return false
}

// idxCombiningAbove ports tokenString.indexOf(ch) > 0 (Java UTF-16 code-unit index).
func idxCombiningAbove(s string, ch rune) bool {
	u16 := 0
	for _, r := range s {
		if r == ch {
			return u16 > 0
		}
		if r > 0xFFFF {
			u16 += 2
		} else {
			u16++
		}
	}
	return false
}

var (
	toLatMap = map[rune]rune{}
	toCyrMap = map[rune]rune{}
)

func init() {
	// Java: cyrChars / latChars static maps
	cyrChars := []rune("аеіїкморстухАВЕІКМНОРСТУХ")
	latChars := []rune("aeiïkmopctyxABEIKMHOPCTYX")
	for i := range cyrChars {
		toLatMap[cyrChars[i]] = latChars[i]
		toCyrMap[latChars[i]] = cyrChars[i]
	}
}

// umlauts / umlautsReplace — Java parallel arrays for toCyrillic.
var (
	umlauts        = []string{"á", "Á", "é", "É", "í", "Í", "ḯ", "Ḯ", "ó", "Ó", "ú", "ý", "Ý"}
	umlautsReplace = []string{"а́", "А́", "е́", "Е́", "і́", "І́", "ї́", "Ї́", "о́", "О́", "и́", "у́", "У́"}
)

// toCyrillic ports MixedAlphabetsRule.toCyrillic.
func toCyrillic(word string) string {
	// Java iterates Map entrySet — order not specified; char replace is commutative for 1:1 map.
	var b strings.Builder
	for _, r := range word {
		if c, ok := toCyrMap[r]; ok {
			b.WriteRune(c)
		} else {
			b.WriteRune(r)
		}
	}
	s := b.String()
	for i := range umlauts {
		s = strings.ReplaceAll(s, umlauts[i], umlautsReplace[i])
	}
	return s
}

// toLatin ports MixedAlphabetsRule.toLatin.
func toLatin(word string) string {
	var b strings.Builder
	for _, r := range word {
		if c, ok := toLatMap[r]; ok {
			b.WriteRune(c)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
