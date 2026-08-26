package ru

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// INNNumberFilter ports org.languagetool.rules.ru.INNNumberFilter.
// Keeps the match when the INN checksum is invalid; suppresses (nil) when valid,
// non-digit, or wrong length (Java default / non-match / catch → null).
type INNNumberFilter struct{}

func NewINNNumberFilter() *INNNumberFilter {
	return &INNNumberFilter{}
}

// Java: private static final Pattern DIGIT_SYMBOL_PATTERN = Pattern.compile("(\\d*)");
// Matcher.matches() requires a full match → equivalent to ^(\d*)$.
var digitSymbolPattern = regexp.MustCompile(`^(\d*)$`)

// AcceptRuleMatch ports INNNumberFilter.acceptRuleMatch:
// getRequired("inn"); if digit match and length 10/12 with bad checksum → match;
// valid checksum / non-digit / other length → null.
func (f *INNNumberFilter) AcceptRuleMatch(match *rules.RuleMatch, args map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	innNumberString := patterns.GetRequired("inn", args)
	if f.checksumInvalid(innNumberString) {
		return match
	}
	return nil
}

// IsInvalid reports whether a 10/12-digit INN fails the Java checksum
// (true ⇔ AcceptRuleMatch would keep the match). Non-digit / wrong-length → false.
func (f *INNNumberFilter) IsInvalid(inn string) bool {
	return f.checksumInvalid(inn)
}

// checksumInvalid mirrors Java acceptRuleMatch body: true only when length is 10 or 12
// and the check digit(s) do not match the weighted sum % 11 (with kz>9 → kz-10).
func (f *INNNumberFilter) checksumInvalid(innNumberString string) bool {
	if !digitSymbolPattern.MatchString(innNumberString) {
		// Java: matcher does not match → return null (not invalid for this filter)
		return false
	}
	// Java: String.valueOf(...).chars().map(Character::getNumericValue).toArray()
	intTab := make([]int, len(innNumberString))
	for i := 0; i < len(innNumberString); i++ {
		// After ^(\d*)$ match, only ASCII '0'..'9' (or empty).
		intTab[i] = int(innNumberString[i] - '0')
	}
	switch len(intTab) {
	case 10:
		// Weights 2,4,10,3,5,9,4,6,8 for digits 0..8; check digit[9]
		kz1 := (intTab[0]*2 + intTab[1]*4 + intTab[2]*10 + intTab[3]*3 + intTab[4]*5 +
			intTab[5]*9 + intTab[6]*4 + intTab[7]*6 + intTab[8]*8) % 11
		if kz1 > 9 {
			kz1 = kz1 - 10
		}
		return intTab[9] != kz1
	case 12:
		// kz1 weights 7,2,4,10,3,5,9,4,6,8 for 0..9; check digit[10]
		kz1 := (intTab[0]*7 + intTab[1]*2 + intTab[2]*4 + intTab[3]*10 + intTab[4]*3 +
			intTab[5]*5 + intTab[6]*9 + intTab[7]*4 + intTab[8]*6 + intTab[9]*8) % 11
		// kz2 weights 3,7,2,4,10,3,5,9,4,6,8 for 0..10; check digit[11]
		kz2 := (intTab[0]*3 + intTab[1]*7 + intTab[2]*2 + intTab[3]*4 + intTab[4]*10 +
			intTab[5]*3 + intTab[6]*5 + intTab[7]*9 + intTab[8]*4 + intTab[9]*6 + intTab[10]*8) % 11
		if kz1 > 9 {
			kz1 = kz1 - 10
		}
		if kz2 > 9 {
			kz2 = kz2 - 10
		}
		return !(intTab[10] == kz1 && intTab[11] == kz2)
	default:
		return false
	}
}
