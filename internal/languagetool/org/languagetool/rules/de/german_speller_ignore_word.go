package de

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
)

// List-form ignoreWord extras from GermanSpellerRule.ignoreWord(List, idx).
// CompoundTokenize optional for hanging-hyphen isCompound; without it only hyphen / SPECIAL_CASE_THIRD.

// adjSuffix ports GermanSpellerRule.adjSuffix (adjective compound endings).
const adjSuffix = `(affin|basiert|konform|widrig|fähig|haltig|bedingt|gerecht|würdig|relevant|` +
	`übergreifend|tauglich|untauglich|artig|bezogen|orientiert|fremd|liebend|hassend|bildend|hemmend|abhängig|zentriert|` +
	`förmig|mäßig|pflichtig|ähnlich|spezifisch|verträglich|technisch|typisch|frei|arm|freundlich|feindlich|gemäß|neutral|seitig|begeistert|geeignet|ungeeignet|berechtigt|sicher|süchtig|resistent|verachtend|schädigend)`

var (
	reMissingAdjPattern = regexp.MustCompile(`^[a-zöäüß]{3,25}` + adjSuffix + `(?:er|es|en|em|e)?$`)
	reAdjSuffixStrip    = regexp.MustCompile(adjSuffix + `(?:er|es|en|em|e)?$`)
	reSpecialCase       = regexp.MustCompile(`^.{3,25}(?:tum|ing|ling|heit|keit|schaft|ung|ion|tät|at|um)$`)
	reSpecialCaseWithS  = regexp.MustCompile(`^.{3,25}(?:tum|ing|ling|heit|keit|schaft|ung|ion|tät|at|um)s$`)
	reSpecialCaseThird  = regexp.MustCompile(`^[A-ZÖÄÜ][a-zöäüß]{2,}(?:ei|öl)$`)
	reMitarbeitend      = regexp.MustCompile(`mitarbeitenden?`)
)

// enumerationConnectors ports the skip set in getWordAfterEnumerationOrNull.
var enumerationConnectors = map[string]struct{}{
	",": {}, "/": {}, "&": {}, "und": {}, "oder": {}, "bzw.": {},
	"beziehungsweise": {}, "sowie": {}, "statt": {},
}

func uncapitalizeFirst(word string) string {
	if word == "" {
		return word
	}
	r := []rune(word)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

// isOnlyNoun ports GermanSpellerRule.isOnlyNoun: every POS tag starts with SUB:.
// Nil TagPOS → false (fail-closed).
func (r *GermanSpellerRule) isOnlyNoun(word string) bool {
	if r == nil || r.TagPOS == nil || word == "" {
		return false
	}
	tags := r.TagPOS(word)
	if len(tags) == 0 {
		return false
	}
	for _, t := range tags {
		if t == "" || !strings.HasPrefix(t, "SUB:") {
			return false
		}
	}
	return true
}

// getWordAfterEnumerationOrNull ports GermanSpellerRule.getWordAfterEnumerationOrNull.
func getWordAfterEnumerationOrNull(words []string, idx int) string {
	for i := idx; i < len(words); i++ {
		w := words[i]
		if strings.HasSuffix(w, "-") {
			continue
		}
		if _, skip := enumerationConnectors[w]; skip {
			continue
		}
		if strings.TrimSpace(w) == "" {
			continue
		}
		return w
	}
	return ""
}

// ignoreByHangingHyphen ports GermanSpellerRule.ignoreByHangingHyphen for
// "Stil- und Grammatikprüfung". Without German compoundTokenizer, isCompound is
// true only when nextWord contains "-" or matches SPECIAL_CASE_THIRD (ei|öl).
func (r *GermanSpellerRule) ignoreByHangingHyphen(words []string, idx int) bool {
	if r == nil || idx < 0 || idx >= len(words) {
		return false
	}
	word := words[idx]
	nextWord := getWordAfterEnumerationOrNull(words, idx+1)
	nextWord = strings.TrimSuffix(nextWord, ".")
	if nextWord == "" {
		return false
	}
	isCompound := strings.Contains(nextWord, "-") || reSpecialCaseThird.MatchString(nextWord)
	if !isCompound && r.CompoundTokenize != nil {
		// Java: compoundTokenizer.tokenize(nextWord).size() > 1
		isCompound = len(r.CompoundTokenize(nextWord)) > 1
	}
	if !isCompound {
		return false
	}
	word = strings.TrimSuffix(word, "-")
	if !FilterDictAvailable() {
		return false
	}
	miss := FilterDictIsMisspelled(word)
	if miss && (r.IgnoreWord(word) || r.IsIgnoredInCompounds(word)) {
		miss = false
	} else if miss && strings.HasSuffix(word, "s") {
		base := strings.TrimSuffix(word, "s")
		if isNeedingFugenS(base) {
			miss = FilterDictIsMisspelled(base)
		}
	}
	return !miss
}

// ignoreMissingAdjCompound ports the missingAdjPattern branch of ignoreWord(List).
func (r *GermanSpellerRule) ignoreMissingAdjCompound(word string) bool {
	if r == nil || !reMissingAdjPattern.MatchString(word) {
		return false
	}
	if !r.IsMisspelled(word) {
		return false
	}
	// firstPart = uppercaseFirstChar(word.replaceFirst(adjSuffix + "(er|…)?", ""))
	stripped := reAdjSuffixStrip.ReplaceAllString(word, "")
	firstPart := uppercaseFirstChar(stripped)
	if firstPart == "" {
		return false
	}
	if !r.IsMisspelled(firstPart) && !reSpecialCase.MatchString(firstPart) &&
		r.isOnlyNoun(firstPart) && !r.IsMisspelled(firstPart+"test") {
		return true
	}
	if strings.HasSuffix(firstPart, "s") {
		withoutS := firstPart[:len(firstPart)-1]
		if !r.IsMisspelled(withoutS) && reSpecialCaseWithS.MatchString(firstPart) &&
			r.isOnlyNoun(withoutS) && !r.IsMisspelled(firstPart+"test") {
			return true
		}
	}
	return false
}

// ignoreMitarbeitende ports the mitarbeitende(n) → mitarbeiter hunspell check.
func (r *GermanSpellerRule) ignoreMitarbeitende(word string) bool {
	if r == nil {
		return false
	}
	if !strings.HasSuffix(word, "mitarbeitende") && !strings.HasSuffix(word, "mitarbeitenden") {
		return false
	}
	if !FilterDictAvailable() {
		return false
	}
	loc := reMitarbeitend.FindStringIndex(word)
	if loc == nil {
		return false
	}
	replaced := word[:loc[0]] + "mitarbeiter" + word[loc[1]:]
	return !FilterDictIsMisspelled(replaced)
}

// IgnoreWordAt ports GermanSpellerRule.ignoreWord(List<String>, int).
// Used by Match-path spelling; string IgnoreWord remains SpellingCheckRule.ignoreWord.
func (r *GermanSpellerRule) IgnoreWordAt(words []string, idx int) bool {
	if r == nil || idx < 0 || idx >= len(words) {
		return false
	}
	word := words[idx]
	if len([]rune(word)) > GermanSpellerMaxTokenLength {
		return true
	}
	// multi-token IGNORE_SPELLING phrases (any covered position)
	if r.isInIgnorePhrase(words, idx) {
		return true
	}
	ignore := r.IgnoreWord(word)
	ignoreUncapitalizedWord := false
	if !ignore && idx == 0 {
		ignoreUncapitalizedWord = r.IgnoreWord(uncapitalizeFirst(words[0]))
	}
	ignoreByHyphen := false
	ignoreBulletPointCase := false
	if !ignoreUncapitalizedWord {
		// Google Docs list items: empty token + uppercased word that is only OK in lowercase
		ignoreBulletPointCase = !ignore && idx == 1 && words[0] == "" &&
			startsWithUppercase(word) &&
			r.IsMisspelled(word) &&
			!r.IsMisspelled(strings.ToLower(word))
	}
	ignoreHyphenatedCompound := false
	if !ignore && !ignoreUncapitalizedWord {
		if strings.Contains(word, "-") {
			if idx > 0 && words[idx-1] == "" &&
				(strings.HasPrefix(word, "stel-") || strings.HasPrefix(word, "tel-")) {
				// '100stel-Millimeter' / '5tel-Gramm'
				after := word
				if i := strings.Index(word, "-"); i >= 0 && i+1 < len(word) {
					after = word[i+1:]
				}
				return !r.IsMisspelled(after)
			}
			ignoreByHyphen = strings.HasSuffix(word, "-") && r.ignoreByHangingHyphen(words, idx)
		}
		ignoreHyphenatedCompound = !ignoreByHyphen && r.IgnoreCompoundWithIgnoredWord(word)
	}
	if spelling.GetSuffixPattern().MatchString(word) {
		return true
	}
	if r.ignoreMissingAdjCompound(word) {
		return true
	}
	if r.ignoreMitarbeitende(word) {
		return true
	}
	if (idx+1 < len(words) && (strings.HasSuffix(word, ".mp") || strings.HasSuffix(word, ".woff")) && words[idx+1] == "") ||
		(idx > 0 && words[idx-1] == "" && (word == "sat" || word == "stel" || word == "tel" || word == "stels" || word == "tels")) {
		return true
	}
	return ignore || ignoreUncapitalizedWord || ignoreBulletPointCase || ignoreByHyphen ||
		ignoreHyphenatedCompound || r.IgnoreElative(word)
}
