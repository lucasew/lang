package ru

import (
	"regexp"
	"unicode"
)

// INNNumberFilter ports org.languagetool.rules.ru.INNNumberFilter.
// Returns true when the INN checksum is invalid (match should be kept).
type INNNumberFilter struct{}

func NewINNNumberFilter() *INNNumberFilter {
	return &INNNumberFilter{}
}

var digitOnly = regexp.MustCompile(`^\d+$`)

// IsInvalid returns true if the INN number fails checksum validation.
// Non-digit / wrong-length inputs return false (suppress; other rules handle them).
func (f *INNNumberFilter) IsInvalid(inn string) bool {
	if !digitOnly.MatchString(inn) {
		return false
	}
	digits := make([]int, len(inn))
	for i, r := range inn {
		if !unicode.IsDigit(r) {
			return false
		}
		digits[i] = int(r - '0')
	}
	switch len(digits) {
	case 10:
		kz1 := (digits[0]*2 + digits[1]*4 + digits[2]*10 + digits[3]*3 + digits[4]*5 +
			digits[5]*9 + digits[6]*4 + digits[7]*6 + digits[8]*8) % 11
		if kz1 > 9 {
			kz1 -= 10
		}
		return digits[9] != kz1
	case 12:
		kz1 := (digits[0]*7 + digits[1]*2 + digits[2]*4 + digits[3]*10 + digits[4]*3 +
			digits[5]*5 + digits[6]*9 + digits[7]*4 + digits[8]*6 + digits[9]*8) % 11
		kz2 := (digits[0]*3 + digits[1]*7 + digits[2]*2 + digits[3]*4 + digits[4]*10 +
			digits[5]*3 + digits[6]*5 + digits[7]*9 + digits[8]*4 + digits[9]*6 + digits[10]*8) % 11
		if kz1 > 9 {
			kz1 -= 10
		}
		if kz2 > 9 {
			kz2 -= 10
		}
		return !(digits[10] == kz1 && digits[11] == kz2)
	default:
		return false
	}
}
