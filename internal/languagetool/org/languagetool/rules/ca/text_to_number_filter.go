package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// TextToNumberFilter ports org.languagetool.rules.ca.TextToNumberFilter.
type TextToNumberFilter struct {
	*rules.TextToNumberFilter
}

// NewTextToNumberFilter builds the Catalan number-word tables.
func NewTextToNumberFilter() *TextToNumberFilter {
	numbers := map[string]float64{
		"zero": 0, "mig": 0.5,
		"un": 1, "u": 1, "una": 1,
		"dos": 2, "dues": 2, "tres": 3, "quatre": 4, "cinc": 5,
		"sis": 6, "set": 7, "vuit": 8, "huit": 8, "nou": 9, "deu": 10,
		"onze": 11, "dotze": 12, "tretze": 13, "catorze": 14, "quinze": 15, "setze": 16,
		"disset": 17, "desset": 17, "dèsset": 17,
		"divuit": 18, "devuit": 18, "díhuit": 18,
		"dinou": 19, "denou": 19, "dènou": 19, "dèneu": 19,
		"vint": 20, "trenta": 30, "quaranta": 40, "cinquanta": 50,
		"seixanta": 60, "setanta": 70, "vuitanta": 80, "huitanta": 80, "noranta": 90,
	}
	multipliers := map[string]float64{
		"cent": 100, "cents": 100,
		"mil":   1000,
		"milió": 1_000_000, "milions": 1_000_000,
		"bilió": 10e12, "bilions": 10e12,
		"trilió": 10e18, "trilions": 10e18,
	}
	base := &rules.TextToNumberFilter{
		Numbers:     numbers,
		Multipliers: multipliers,
		IsComma: func(s string) bool {
			s = strings.ToLower(s)
			return s == "comma" || s == "coma"
		},
		IsPercentage: func(tokens []string, i int) bool {
			if i <= 0 || i >= len(tokens) {
				return false
			}
			return tokens[i] == "cent" && strings.ToLower(tokens[i-1]) == "per"
		},
		Tokenize: func(s string) []string {
			return strings.Split(s, "-")
		},
		FormatResult: func(s string) string {
			return strings.ReplaceAll(s, ".", ",")
		},
	}
	return &TextToNumberFilter{TextToNumberFilter: base}
}
