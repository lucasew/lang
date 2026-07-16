package uk

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// MixedAlphabetsRule ports org.languagetool.rules.uk.MixedAlphabetsRule.
type MixedAlphabetsRule struct {
	Messages map[string]string
}

func NewMixedAlphabetsRule(messages map[string]string) *MixedAlphabetsRule {
	return &MixedAlphabetsRule{Messages: messages}
}

func (r *MixedAlphabetsRule) GetID() string { return "UK_MIXED_ALPHABETS" }

var (
	likelyLatinNumber   = regexp.MustCompile(`[XVIХІ]{2,8}(-[а-яіїє]{1,3})?`)
	latinNumberWithCyr  = regexp.MustCompile(`(Х{1,3}І{1,3}|І{1,3}Х{1,3}|Х{2,3}|І{2,3})(-[а-яіїє]{1,4})?`)
	mixedAlphabets      = regexp.MustCompile(`.*([a-zA-ZïáÁéÉíÍḯḮóÓúýÝ]'?[а-яіїєґА-ЯІЇЄҐ]|[а-яіїєґА-ЯІЇЄҐ]'?[a-zA-ZïáÁéÉíÍḯḮóÓúýÝ]).*`)
	cyrillicOnly        = regexp.MustCompile(`.*[бвгґдєжзийїлнпфцчшщьюяБГҐДЄЖЗИЙЇЛПФЦЧШЩЬЮЯ].*`)
	latinOnly           = regexp.MustCompile(`.*[bdfghjlqrstvzDFGJLNQRSUVZ].*`)
	commonCyrLetters    = regexp.MustCompile(`^[АВЕІКОРСТУХ]+$`)
	cyrillicFirstLetter = regexp.MustCompile(`^[а-яіїєґА-ЯІЇЄҐ]`)
	singleLatinIYA      = regexp.MustCompile(`^[iya]$`)
)

func (r *MixedAlphabetsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var ruleMatches []*rules.RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		tokenReadings := tokens[i]
		tokenString := tokenReadings.GetToken() // clean token stand-in
		endReadings := tokenReadings
		// Join "І" + "." (Java treats "І." as one token for Roman numeral I.)
		if tokenString == "І" && i+1 < len(tokens) && tokens[i+1].GetToken() == "." && !tokens[i+1].IsWhitespaceBefore() {
			tokenString = "І."
			endReadings = tokens[i+1]
		}

		// 1-letter latin i/y/a before cyrillic (except formulas with x/b/B)
		if i < len(tokens)-1 &&
			(singleLatinIYA.MatchString(tokenString) || (tokenString == "A" && i == 1)) &&
			cyrillicFirstLetter.MatchString(tokens[i+1].GetToken()) &&
			!hasAnyToken(tokens, "x", "b", "B") {
			msg := "Вжито латинську «" + tokenString + "» замість кириличної"
			ruleMatches = append(ruleMatches, r.createRuleMatch(tokenReadings, []string{toCyrillic(tokenString)}, msg, sentence))
		} else if tokenString == "І" && likelyBadLatinI(tokens, i) {
			msg := "Вжито кириличну літеру замість латинської"
			ruleMatches = append(ruleMatches, r.createRuleMatch(tokenReadings, []string{toLatin(tokenString)}, msg, sentence))
		} else if tokenString == "І." &&
			i > 1 && tokens[i-1].GetToken() != "Тому" && tokens[i-1].GetToken() != "Франко" &&
			isCapitalizedUk(tokens[i-1].GetToken()) {
			// fname stand-in: capitalized previous word (Петро/Миколая І.)
			msg := "Вжито кириличну літеру замість латинської"
			ruleMatches = append(ruleMatches, r.createRuleMatchSpan(tokenReadings, endReadings, []string{toLatin(tokenString)}, msg, sentence))
			if endReadings != tokenReadings {
				i++ // skip joined period
			}
		} else if commonCyrLetters.MatchString(tokenString) {
			// prev lemma гепатит|група|турнір — surface: prev ends with those stems or is "група"/"гепатиту"
			prev := tokens[i-1].GetToken()
			if isHepatitisGroupTournament(prev) {
				msg := "Вжито кириличну літеру замість латинської"
				ruleMatches = append(ruleMatches, r.createRuleMatch(tokenReadings, []string{toLatin(tokenString)}, msg, sentence))
			}
		}

		// Degree + Cyrillic С (may be one token "6°С" after tokenizer)
		if strings.Contains(tokenString, "°С") {
			msg := "Вжито кириличну літеру замість латинської"
			ruleMatches = append(ruleMatches, r.createRuleMatch(tokenReadings, []string{"C"}, msg, sentence))
		}
		if utf8.RuneCountInString(tokenString) < 2 {
			if tokenString == "°" && i < len(tokens)-1 && tokens[i+1].GetToken() == "С" {
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
			if (utf8.RuneCountInString(tokenString) > 2 && !cyrillicOnly.MatchString(tokenString)) ||
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
			converted := adjustForInvalidSuffix(toLatinLeftOnly(tokenString))
			msg := "Вжито кириличні літери замість латинських на позначення римської цифри"
			msg = adjustForInvalidSuffixMsg(tokenString, msg)
			ruleMatches = append(ruleMatches, r.createRuleMatch(tokenReadings, []string{converted}, msg, sentence))
		}

		if strings.ContainsRune(tokenString, '\u0306') || strings.ContainsRune(tokenString, '\u0308') {
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

func (r *MixedAlphabetsRule) createRuleMatch(readings *languagetool.AnalyzedTokenReadings, replacements []string, msg string, sentence *languagetool.AnalyzedSentence) *rules.RuleMatch {
	return r.createRuleMatchSpan(readings, readings, replacements, msg, sentence)
}

func (r *MixedAlphabetsRule) createRuleMatchSpan(from, to *languagetool.AnalyzedTokenReadings, replacements []string, msg string, sentence *languagetool.AnalyzedSentence) *rules.RuleMatch {
	rm := rules.NewRuleMatch(r, sentence, from.GetStartPos(), to.GetEndPos(), msg)
	rm.ShortMessage = "Мішанина розкладок"
	rm.SetSuggestedReplacements(replacements)
	return rm
}

func likelyBadLatinI(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if i <= 1 {
		return false
	}
	prev := tokens[i-1].GetToken()
	if isCapitalizedUk(prev) {
		return true
	}
	// prep stand-in: short all-upper-or-title prepositions without full tagger
	if isLikelyPrep(prev) && i < len(tokens)-1 && !isAllUppercaseUk(tokens[i+1].GetToken()) {
		return true
	}
	if i < len(tokens)-1 {
		next := tokens[i+1].GetToken()
		if next == "ст." || next == "тис." {
			return true
		}
		switch next {
		case "квартал", "півріччя", "тисячоліття", "половина":
			return true
		}
	}
	return false
}

func isLikelyPrep(s string) bool {
	switch strings.ToLower(s) {
	case "у", "в", "на", "за", "до", "з", "із", "про", "від":
		return true
	}
	return false
}

func isHepatitisGroupTournament(prev string) bool {
	pl := strings.ToLower(prev)
	return strings.Contains(pl, "гепатит") || pl == "група" || strings.Contains(pl, "турнір")
}

func isCapitalizedUk(s string) bool {
	if s == "" {
		return false
	}
	// Require a multi-letter name (not "В." initials).
	letters := 0
	for _, c := range s {
		if unicode.IsLetter(c) {
			letters++
		}
	}
	if letters < 2 {
		return false
	}
	r, size := utf8.DecodeRuneInString(s)
	if !unicode.IsUpper(r) {
		return false
	}
	for _, c := range s[size:] {
		if unicode.IsLetter(c) && !unicode.IsLower(c) {
			return false
		}
	}
	return true
}

func isAllUppercaseUk(s string) bool {
	hasLetter := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}

func hasAnyToken(tokens []*languagetool.AnalyzedTokenReadings, vals ...string) bool {
	set := map[string]bool{}
	for _, v := range vals {
		set[v] = true
	}
	for _, t := range tokens {
		if set[t.GetToken()] {
			return true
		}
	}
	return false
}

func adjustForInvalidSuffix(tokenString string) string {
	if strings.Contains(tokenString, "-") {
		re := regexp.MustCompile(`-[а-яіїє]{1,4}$`)
		tokenString = re.ReplaceAllString(tokenString, "")
	}
	return tokenString
}

func adjustForInvalidSuffixMsg(tokenString, msg string) string {
	if strings.Contains(tokenString, "-") {
		if matched, _ := regexp.MatchString(`^[IVXІХ]+-[а-яіїє]{1,4}$`, tokenString); matched {
			msg += ". Також: до римських цифр букви не дописуються."
		}
	}
	return msg
}

func toLatinLeftOnly(tokenString string) string {
	parts := strings.SplitN(tokenString, "-", 2)
	right := ""
	if len(parts) > 1 {
		right = "-" + parts[1]
	}
	return toLatin(parts[0]) + right
}

var (
	toLatMap = map[rune]rune{}
	toCyrMap = map[rune]rune{}
)

func init() {
	cyrChars := []rune("аеіїкморстухАВЕІКМНОРСТУХ")
	latChars := []rune("aeiïkmopctyxABEIKMHOPCTYX")
	for i := range cyrChars {
		toLatMap[cyrChars[i]] = latChars[i]
		toCyrMap[latChars[i]] = cyrChars[i]
	}
}

var (
	umlauts        = []string{"á", "Á", "é", "É", "í", "Í", "ḯ", "Ḯ", "ó", "Ó", "ú", "ý", "Ý"}
	umlautsReplace = []string{"а́", "А́", "е́", "Е́", "і́", "І́", "ї́", "Ї́", "о́", "О́", "и́", "у́", "У́"}
)

func toCyrillic(word string) string {
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
