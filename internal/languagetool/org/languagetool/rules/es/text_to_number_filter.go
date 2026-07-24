package es

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// TextToNumberFilter ports org.languagetool.rules.es.TextToNumberFilter.
type TextToNumberFilter struct {
	*rules.TextToNumberFilter
}

// NewTextToNumberFilter builds the Spanish number-word tables.
func NewTextToNumberFilter() *TextToNumberFilter {
	numbers := map[string]float64{
		"cero": 0, "medio": 0.5,
		"un": 1, "uno": 1, "una": 1,
		"dos": 2, "tres": 3, "cuatro": 4, "cinco": 5,
		"seis": 6, "siete": 7, "ocho": 8, "nueve": 9,
		"diez": 10, "once": 11, "doce": 12, "trece": 13, "catorce": 14, "quince": 15,
		"dieciséis": 16, "diecisiete": 17, "dieciocho": 18, "diecinueve": 19,
		"veinte": 20, "veintiuno": 21, "veintidós": 22, "veintitrés": 23,
		"veinticuatro": 24, "veinticinco": 25, "veintiséis": 26, "veintisiete": 27,
		"veintiocho": 28, "veintinueve": 29,
		"treinta": 30, "cuarenta": 40, "cincuenta": 50, "sesenta": 60,
		"setenta": 70, "ochenta": 80, "noventa": 90,
		"cien": 100, "ciento": 100,
		"doscientos": 200, "trescientos": 300, "cuatrocientos": 400, "quinientos": 500,
		"seiscientos": 600, "setecientos": 700, "ochocientos": 800, "novecientos": 900,
		"doscientas": 200, "trescientas": 300, "cuatrocientas": 400, "quinientas": 500,
		"seiscientas": 600, "setecientas": 700, "ochocientas": 800, "novecientas": 900,
	}
	multipliers := map[string]float64{
		"mil":    1000,
		"millón": 1_000_000, "millones": 1_000_000,
		// Match Java 10E12 / 10E18 literals (1e13 / 1e19).
		"billón": 10e12, "billones": 10e12,
		"trillón": 10e18, "trillones": 10e18,
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
			return tokens[i] == "ciento" && strings.ToLower(tokens[i-1]) == "por"
		},
	}
	return &TextToNumberFilter{TextToNumberFilter: base}
}
