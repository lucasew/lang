package uk

import "regexp"

// LetterEndingForNumericHelper ports
// org.languagetool.tagging.uk.LetterEndingForNumericHelper (subset).
// Maps Ukrainian numeric adjective/noun letter endings to POS case tags.
type LetterEndingForNumericHelper struct{}

type regexToCaseList struct {
	re   *regexp.Regexp // nil = always
	tags []string
}

var numrAdjEndingMap = map[string][]regexToCaseList{
	"й":  {{tags: []string{":m:v_naz", ":m:v_zna:rinanim", ":f:v_dav", ":f:v_mis"}}},
	"ий": {{tags: []string{":m:v_naz", ":m:v_zna:rinanim"}}},
	"а":  {{tags: []string{":f:v_naz"}}},
	"у":  {{tags: []string{":f:v_zna"}}},
	"го": {{tags: []string{":m:v_rod", ":m:v_zna:ranim", ":n:v_rod"}}},
	"му": {{tags: []string{":m:v_dav", ":m:v_mis", ":n:v_dav", ":n:v_mis"}}},
	"м":  {{tags: []string{":m:v_oru", ":n:v_oru", ":p:v_dav"}}},
	"им": {{tags: []string{":m:v_oru", ":n:v_oru", ":p:v_dav"}}},
	"ої": {{tags: []string{":f:v_rod"}}},
	"ї":  {{tags: []string{":f:v_rod"}}},
	"ою": {{tags: []string{":f:v_oru"}}},
}

// TagsForAdjEnding returns case tags for a letter ending (e.g. "й", "а").
// number is the numeric base (for regex-conditioned endings); empty matches always rules only.
func (LetterEndingForNumericHelper) TagsForAdjEnding(ending, number string) []string {
	lists, ok := numrAdjEndingMap[ending]
	if !ok {
		return nil
	}
	var out []string
	for _, item := range lists {
		if item.re == nil || (number != "" && item.re.MatchString(number)) {
			out = append(out, item.tags...)
		}
	}
	return out
}

// HasKnownEnding reports whether ending is in the adj map.
func (LetterEndingForNumericHelper) HasKnownEnding(ending string) bool {
	_, ok := numrAdjEndingMap[ending]
	return ok
}

// CasesForNumericEnding returns POS case suffixes for number+letter ending pairs.
func CasesForNumericEnding(number, ending string) []string {
	var h LetterEndingForNumericHelper
	return h.TagsForAdjEnding(ending, number)
}
