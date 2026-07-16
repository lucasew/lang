package pt

import (
	"strconv"
	"strings"
)

// RomanNumeralFilter ports org.languagetool.rules.pt.RomanNumeralFilter
// without the full synthesizer/Soros pipeline.
type RomanNumeralFilter struct{}

func NewRomanNumeralFilter() *RomanNumeralFilter {
	return &RomanNumeralFilter{}
}

// Suggest returns the Roman form of an arabic numeral string (e.g. "2024" → "MMXXIV").
// Empty input or non-numeric yields "".
func (f *RomanNumeralFilter) Suggest(arabicSource string) string {
	return ToRoman(arabicSource)
}

// ToRoman converts a decimal integer string to uppercase Roman numerals.
// Supports 1..3999; out of range returns "".
func ToRoman(arabic string) string {
	arabic = strings.TrimSpace(arabic)
	n, err := strconv.Atoi(arabic)
	if err != nil || n <= 0 || n > 3999 {
		return ""
	}
	vals := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	syms := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
	var b strings.Builder
	for i, v := range vals {
		for n >= v {
			b.WriteString(syms[i])
			n -= v
		}
	}
	return b.String()
}
