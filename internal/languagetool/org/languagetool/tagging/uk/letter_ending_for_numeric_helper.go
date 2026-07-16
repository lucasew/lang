package uk

import "regexp"

// LetterEndingForNumericHelper ports a subset of numeric adjective letter endings.
// Maps ending → POS case fragments (always-applied entries only for simplicity).
type letterEndingCases struct {
	// if re is nil, always apply cases
	re    *regexp.Regexp
	cases []string
}

var numrAdjEndingMap = map[string][]letterEndingCases{
	"й":  {{cases: []string{":m:v_naz", ":m:v_zna:rinanim", ":f:v_dav", ":f:v_mis"}}},
	"ий": {{cases: []string{":m:v_naz", ":m:v_zna:rinanim"}}},
	"а":  {{cases: []string{":f:v_naz"}}},
	"го": {{cases: []string{":m:v_rod", ":m:v_zna:ranim", ":n:v_rod"}}},
	"им": {{cases: []string{":m:v_oru", ":n:v_oru", ":p:v_dav"}}},
	"ої": {{cases: []string{":f:v_rod"}}},
	"ою": {{cases: []string{":f:v_oru"}}},
}

// CasesForNumericEnding returns POS case fragments for a letter ending after a number.
// number is the digit stem (e.g. "1", "42"); ending is without hyphen (e.g. "й").
func CasesForNumericEnding(number, ending string) []string {
	list, ok := numrAdjEndingMap[ending]
	if !ok {
		return nil
	}
	var out []string
	for _, e := range list {
		if e.re != nil && !e.re.MatchString(number) {
			continue
		}
		out = append(out, e.cases...)
	}
	return out
}
