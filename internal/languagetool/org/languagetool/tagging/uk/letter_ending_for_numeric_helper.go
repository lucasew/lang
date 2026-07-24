package uk

import (
	"regexp"
)

// LetterEndingForNumericHelper ports
// org.languagetool.tagging.uk.LetterEndingForNumericHelper.
// Maps Ukrainian numeric adjective/noun letter endings to POS case tags.
type LetterEndingForNumericHelper struct{}

type regexToCaseList struct {
	// matchNumber returns true when this rule applies; nil = always.
	matchNumber func(number string) bool
	tags        []string
}

func always(tags ...string) regexToCaseList {
	return regexToCaseList{tags: tags}
}

func whenRE(pattern string, tags ...string) regexToCaseList {
	re := regexp.MustCompile(pattern)
	return regexToCaseList{
		matchNumber: func(n string) bool { return re.MatchString(n) },
		tags:        tags,
	}
}

// endsWithDigitNotAfter1: Java-ish ".*(?!<1)D" → ends with D but not 1D (e.g. 7 but not 17).
func endsWithDigitNotAfter1(digit byte) func(string) bool {
	return func(n string) bool {
		if n == "" || n[len(n)-1] != digit {
			return false
		}
		if len(n) >= 2 && n[len(n)-2] == '1' {
			return false
		}
		return true
	}
}

func whenFn(fn func(string) bool, tags ...string) regexToCaseList {
	return regexToCaseList{matchNumber: fn, tags: tags}
}

// Java NUMR_ADJ_ENDING_MAP (full). First matching rule wins (Java getCaseTags).
var numrAdjEndingMap = map[string][]regexToCaseList{
	"й":  {always(":m:v_naz", ":m:v_zna:rinanim", ":f:v_dav", ":f:v_mis")},
	"ий": {always(":m:v_naz", ":m:v_zna:rinanim")},
	"ій": {
		whenRE(`.*([^3]|13)$`, ":f:v_dav", ":f:v_mis"),
		always(":m:v_naz", ":m:v_zna:rinanim", ":f:v_dav", ":f:v_mis"),
	},
	"го": {always(":m:v_rod", ":m:v_zna:ranim", ":n:v_rod")},
	"му": {
		whenFn(endsWithDigitNotAfter1('7'), ":m:v_dav", ":m:v_mis", ":n:v_dav", ":n:v_mis", ":f:v_zna"),
		whenFn(endsWithDigitNotAfter1('8'), ":f:v_zna", ":m:v_dav", ":m:v_mis", ":n:v_dav", ":n:v_mis"),
		always(":m:v_dav", ":m:v_mis", ":n:v_dav", ":n:v_mis"),
	},
	"ма": {whenFn(endsWithDigitNotAfter1('7'), ":f:v_naz"), whenFn(endsWithDigitNotAfter1('8'), ":f:v_naz")},
	"м":  {always(":m:v_oru", ":n:v_oru", ":p:v_dav")},
	"им": {always(":m:v_oru", ":n:v_oru", ":p:v_dav")},
	"ім": {
		whenFn(endsWithDigitNotAfter1('3'), ":m:v_oru", ":m:v_mis", ":n:v_oru", ":n:v_mis"),
		always(":m:v_mis", ":n:v_oru", ":n:v_mis"),
	},
	"а":  {always(":f:v_naz")},
	"ва": {always(":f:v_naz")},
	"ша": {always(":f:v_naz")},
	"га": {always(":f:v_naz")},
	"тя": {always(":f:v_naz")},
	"я":  {whenFn(endsWithDigitNotAfter1('3'), ":f:v_naz")},
	"та": {always(":f:v_naz")},
	"ї":  {always(":f:v_rod")},
	"ої": {always(":f:v_rod")},
	"у":  {always(":f:v_zna")},
	"шу": {always(":f:v_zna")},
	"гу": {always(":f:v_zna")},
	"ту": {always(":f:v_zna")},
	"тю": {always(":f:v_zna")},
	"ою": {always(":f:v_oru")},
	"ю": {
		whenRE(`.*([^3]|13)$`, ":f:v_oru"),
		always(":f:v_zna", ":f:v_oru"),
	},
	"е":  {always(":n:v_naz", ":n:v_zna")},
	"є":  {always(":n:v_naz", ":n:v_zna")},
	"ше": {always(":n:v_naz", ":n:v_zna")},
	"ге": {always(":n:v_naz", ":n:v_zna")},
	"тє": {always(":n:v_naz", ":n:v_zna")},
	"те": {always(":n:v_naz", ":n:v_zna")},
	"ме": {whenFn(endsWithDigitNotAfter1('7'), ":n:v_naz", ":n:v_zna"), whenFn(endsWithDigitNotAfter1('8'), ":n:v_naz", ":n:v_zna")},
	"і":  {always(":p:v_naz", ":p:v_zna:rinanim")},
	"ті": {always(":p:v_naz", ":p:v_zna:rinanim")},
	"ні": {always(":p:v_naz", ":p:v_zna:rinanim")},
	"ми": {always(":p:v_oru")},
	"х":  {always(":p:v_rod", ":p:v_zna:ranim", ":p:v_mis")},
	"их": {always(":p:v_rod", ":p:v_zna:ranim", ":p:v_mis")},
	"ві": {
		whenRE(`.*40$`, ":p:v_naz", ":p:v_zna:rinanim"),
		whenRE(`.*%$`, ":p:v_naz", ":p:v_zna:rinanim"),
	},
	// bad
	"тій": {
		whenRE(`.*([^3]|13)$`, ":f:v_dav:bad", ":f:v_mis:bad"),
		always(":m:v_naz:bad", ":m:v_zna:rinanim:bad", ":f:v_dav:bad", ":f:v_mis:bad"),
	},
	"мій":   {always(":f:v_dav:bad", ":f:v_mis:bad")},
	"мою":   {always(":f:v_oru:bad")},
	"тою":   {always(":f:v_oru:bad")},
	"тої":   {always(":f:v_rod:bad")},
	"того":  {always(":m:v_rod:bad", ":n:v_rod:bad")},
	"тього": {always(":m:v_rod:bad", ":n:v_rod:bad")},
	"тому":  {always(":m:v_dav:bad", ":m:v_mis:bad", ":n:v_rod:bad", ":n:v_mis:bad")},
	"тьому": {always(":m:v_dav:bad", ":m:v_mis:bad", ":n:v_rod:bad", ":n:v_mis:bad")},
	"тими":  {always(":p:v_oru:bad")},
	"тім":   {always(":m:v_mis:bad", ":n:v_mis:bad")},
	"мої":   {always(":f:v_rod:bad")},
	"тий":   {always(":m:v_naz:bad", ":m:v_zna:rinanim:bad")},
	"мий":   {always(":m:v_naz:bad", ":m:v_zna:rinanim:bad")},
	"тих":   {always(":p:v_rod:bad", ":p:v_mis:bad")},
	"ого":   {always(":m:v_rod:bad", ":m:v_zna:ranim:bad", ":n:v_rod:bad")},
	"ому":   {always(":m:v_dav:bad", ":m:v_mis:bad", ":n:v_dav:bad", ":n:v_mis:bad")},
	"тим":   {always(":m:v_oru:bad", ":n:v_oru:bad", ":p:v_dav:bad")},
	"ома":   {always(":f:v_naz:bad", ":p:v_oru:bad")},
	"ший":   {always(":m:v_naz:bad", ":m:v_zna:rinanim:bad")},
	"гій":   {always(":f:v_mis:bad", ":f:v_dav:bad")},
}

// Java NUMR_NOUN_ENDING_MAP (full). First matching rule wins.
var numrNounEndingMap = map[string][]regexToCaseList{
	"ти": {whenRE(`.*([0569]|1[0-9])$`, ":p:v_rod:bad", ":p:v_dav:bad", ":p:v_mis:bad")},
	"ці": {whenRE(`.*([03456789]|1[0-9])$`, ":f:v_dav:bad", ":f:v_mis:bad")},
	"ма": {whenRE(`.*([023456789]|1[0-9])$`, ":p:v_oru:bad")},
	"ми": {always(":p:v_rod:bad", ":p:v_mis:bad")},
	"ох": {always(":p:v_rod:bad", ":p:v_zna:ranim:bad")},
	"ві": {whenFn(endsWithDigitNotAfter1('2'), ":p:v_naz:bad", ":p:v_zna:rinanim:bad")},
	"ть": {always(":p:v_naz:bad", ":p:v_zna:rinanim:bad")},
	"ка": {always(":f:v_naz:bad")},
}

func firstMatchingTags(number string, lists []regexToCaseList) []string {
	for _, item := range lists {
		if item.matchNumber == nil || item.matchNumber(number) {
			// copy to avoid callers mutating package map slices
			out := make([]string, len(item.tags))
			copy(out, item.tags)
			return out
		}
	}
	return nil
}

// FindTagsAdj ports LetterEndingForNumericHelper.findTagsAdj.
func FindTagsAdj(leftWord, rightWord string) []string {
	lists, ok := numrAdjEndingMap[rightWord]
	if !ok {
		return nil
	}
	return firstMatchingTags(leftWord, lists)
}

// FindTagsNoun ports LetterEndingForNumericHelper.findTagsNoun.
func FindTagsNoun(leftWord, rightWord string) []string {
	lists, ok := numrNounEndingMap[rightWord]
	if !ok {
		return nil
	}
	return firstMatchingTags(leftWord, lists)
}

// IsPossibleAdjAdjEnding ports LetterEndingForNumericHelper.isPossibleAdjAdjEnding.
func IsPossibleAdjAdjEnding(leftWord, rightWord string) bool {
	_, ok := numrAdjEndingMap[rightWord]
	return ok
}

// TagsForAdjEnding returns case tags for a letter ending (e.g. "й", "а").
// number is the numeric base (for regex-conditioned endings).
func (LetterEndingForNumericHelper) TagsForAdjEnding(ending, number string) []string {
	return FindTagsAdj(number, ending)
}

// HasKnownEnding reports whether ending is in the adj map.
func (LetterEndingForNumericHelper) HasKnownEnding(ending string) bool {
	_, ok := numrAdjEndingMap[ending]
	return ok
}

// CasesForNumericEnding returns POS case suffixes for number+letter ending pairs.
func CasesForNumericEnding(number, ending string) []string {
	return FindTagsAdj(number, ending)
}
