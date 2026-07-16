package pl

import (
	"strconv"
	"strings"
)

// DecadeSpellingFilter ports org.languagetool.rules.pl.DecadeSpellingFilter.
type DecadeSpellingFilter struct{}

func NewDecadeSpellingFilter() *DecadeSpellingFilter {
	return &DecadeSpellingFilter{}
}

// FormatMessage replaces {dekada} and {wiek} in the rule message.
// lata is a 4-digit year-like string: first 2 = century base, last 2 = decade.
// Returns empty string if lata is unparseable.
func (f *DecadeSpellingFilter) FormatMessage(message, lata string) string {
	if len(lata) < 4 {
		return ""
	}
	decade := lata[2:]
	century := lata[:2]
	cent, err := strconv.Atoi(century)
	if err != nil {
		return ""
	}
	msg := strings.ReplaceAll(message, "{dekada}", decade)
	msg = strings.ReplaceAll(msg, "{wiek}", toRoman(cent+1))
	return msg
}

func toRoman(num int) string {
	if num <= 0 {
		return ""
	}
	vals := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	letters := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
	var b strings.Builder
	n := num
	for i, v := range vals {
		for n >= v {
			b.WriteString(letters[i])
			n -= v
		}
	}
	return b.String()
}
