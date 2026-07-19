package pt

import (
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RomanNumeralFilter ports org.languagetool.rules.pt.RomanNumeralFilter.
// Uses standard Roman conversion (1..3999) equivalent to BaseSynthesizer Roman.sor
// for common year/date ranges used in grammar rules.
type RomanNumeralFilter struct{}

func NewRomanNumeralFilter() *RomanNumeralFilter {
	return &RomanNumeralFilter{}
}

// AcceptRuleMatch ports RomanNumeralFilter.acceptRuleMatch.
// Args: arabicSource — decimal numeral string to convert.
func (f *RomanNumeralFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	roman := f.Suggest(arguments["arabicSource"])
	// Java always setSuggestedReplacement (may be empty if Soros fails); keep match.
	match.SetSuggestedReplacement(roman)
	return match
}

// Suggest returns the Roman form of an arabic numeral string (e.g. "2024" → "MMXXIV").
// Empty input or non-numeric yields "".
func (f *RomanNumeralFilter) Suggest(arabicSource string) string {
	return ToRoman(arabicSource)
}

// ToRoman converts a decimal integer string to uppercase Roman numerals.
// Supports 1..3999; out of range returns "" (Java Soros uses overlines above that).
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
