package ca

import (
	"regexp"
	"strings"
)

// ConvertToGenderAndNumberFilter ports POS gender/number split helpers from
// org.languagetool.rules.ca.ConvertToGenderAndNumberFilter.
type ConvertToGenderAndNumberFilter struct{}

func NewConvertToGenderAndNumberFilter() *ConvertToGenderAndNumberFilter {
	return &ConvertToGenderAndNumberFilter{}
}

// GenderAndNumberSplit holds pieces of a Catalan POS tag around gender/number slots.
type GenderAndNumberSplit struct {
	Prefix string
	Suffix string
	Gender string
	Number string
}

var (
	splitGenderNumber          = regexp.MustCompile(`^(N.|A..|V.P..|D..|PX.)(.)(.)(.*)$`)
	splitGenderNumberNoNoun    = regexp.MustCompile(`^(A..|V.P..|D..|PX.)(.)(.)(.*)$`)
	splitGenderNumberAdjective = regexp.MustCompile(`^(A..|V.P..|PX.)(.)(.)(.*)$`)
	postagExceptionsGN         = regexp.MustCompile(`^NP.*$|^AQ0CN0$|^SPS00$|^[CP].*$`)
)

// FormsToIgnore are tokens that stop gender/number expansion.
var FormsToIgnore = map[string]struct{}{
	"mes": {}, "las": {},
}

// SplitGenderAndNumber parses gender/number slots from a POS tag.
// For verbs (V…), gender and number positions are swapped vs nouns/adj.
func SplitGenderAndNumber(pos string) *GenderAndNumberSplit {
	if pos == "" {
		return nil
	}
	m := splitGenderNumber.FindStringSubmatch(pos)
	if m == nil {
		return nil
	}
	res := &GenderAndNumberSplit{
		Prefix: m[1],
		Suffix: m[4],
	}
	g2, g3 := m[2], m[3]
	if strings.HasPrefix(res.Prefix, "V") {
		res.Gender = g3
		res.Number = g2
	} else {
		res.Gender = g2
		res.Number = g3
	}
	return res
}

// DesiredPostag builds a synthesizer postag pattern for desired gender/number.
// For verbs, gender/number are swapped in the tag string as in Java.
func (f *ConvertToGenderAndNumberFilter) DesiredPostag(split *GenderAndNumberSplit, gender, number string) string {
	if split == nil {
		return ""
	}
	g, n := gender, number
	if strings.HasPrefix(split.Prefix, "V") {
		g, n = number, gender
	}
	addGender := "C"
	if strings.HasPrefix(split.Prefix, "DA") {
		addGender = ""
	}
	return split.Prefix + "[" + g + addGender + "]" + "[" + n + "N" + "]" + split.Suffix
}

// ShouldIgnoreForm reports tokens that stop expansion.
func ShouldIgnoreForm(token string) bool {
	_, ok := FormsToIgnore[strings.ToLower(token)]
	return ok
}

// IsPostagException reports POS tags excluded from gender/number rewrite.
func IsPostagException(pos string) bool {
	return postagExceptionsGN.MatchString(pos)
}

// BoToBon special-cases Catalan "bo" → "bon" before nouns (Java synthesize path).
func BoToBon(s string) string {
	if s == "bo" {
		return "bon"
	}
	return s
}

// MatchesSplitGenderNumber reports whether POS matches the main split pattern.
func MatchesSplitGenderNumber(pos string) bool {
	return splitGenderNumber.MatchString(pos)
}

// MatchesAdjectiveSplit reports adjective/participle POS splits.
func MatchesAdjectiveSplit(pos string) bool {
	return splitGenderNumberAdjective.MatchString(pos)
}
