package ca

import (
	"regexp"
	"strings"
)

// SynthesizeWithDAFilter ports surface gender/number + determiner prefixing
// from org.languagetool.rules.ca.SynthesizeWithDAFilter.
// Form synthesis is supplied by the caller via Forms.
type SynthesizeWithDAFilter struct{}

func NewSynthesizeWithDAFilter() *SynthesizeWithDAFilter {
	return &SynthesizeWithDAFilter{}
}

// GenderNumber patterns for matching POS tags to MS/FS/MP/FP.
var genderNumberPatterns = map[string]*regexp.Regexp{
	"MS": regexp.MustCompile(`^(N|A.).[MC][SN].*|V\.P.*SM.`),
	"FS": regexp.MustCompile(`^(N|A.).[FC][SN].*|V\.P.*SF.`),
	"MP": regexp.MustCompile(`^(N|A.).[MC][PN].*|V\.P.*PM.`),
	"FP": regexp.MustCompile(`^(N|A.).[FC][PN].*|V\.P.*PF.`),
}

// GenderNumberFromPOS returns MS/FS/MP/FP for a POS tag, or "".
func GenderNumberFromPOS(pos string) string {
	for _, gn := range []string{"MS", "FS", "MP", "FP"} {
		if genderNumberPatterns[gn].MatchString(pos) {
			return gn
		}
	}
	return ""
}

// PrefixedSuggestion builds determiner/preposition + form for a gender/number.
// preposition is "", "a", "de", or "per" (or first letter thereof).
func (f *SynthesizeWithDAFilter) PrefixedSuggestion(form, genderNumber, preposition string) string {
	det := GetPrepositionAndDeterminer(form, genderNumber, preposition)
	return det + form
}

// FilterForms keeps forms whose POS matches desired gender/number when set.
func (f *SynthesizeWithDAFilter) FilterForms(forms []struct{ Form, POS string }, wantGN string) []string {
	var out []string
	for _, fr := range forms {
		if wantGN != "" {
			if !genderNumberPatterns[wantGN].MatchString(fr.POS) {
				continue
			}
		}
		out = append(out, fr.Form)
	}
	return out
}

// PreferGenderNumber moves forms matching secondGenderNumber (e.g. "MS") earlier.
func PreferGenderNumber(forms []struct{ Form, POS string }, secondGenderNumber string) []struct{ Form, POS string } {
	if secondGenderNumber == "" || len(forms) < 2 {
		return forms
	}
	var preferred, rest []struct{ Form, POS string }
	swap := ""
	if len(secondGenderNumber) >= 2 {
		swap = string(secondGenderNumber[1]) + string(secondGenderNumber[0])
	}
	for i, fr := range forms {
		if i == 0 {
			preferred = append(preferred, fr)
			continue
		}
		if strings.Contains(fr.POS, secondGenderNumber) || (swap != "" && strings.Contains(fr.POS, swap)) {
			preferred = append(preferred, fr)
		} else {
			rest = append(rest, fr)
		}
	}
	return append(preferred, rest...)
}
