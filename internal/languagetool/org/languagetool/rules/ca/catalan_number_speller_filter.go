package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CatalanNumberSpellerFilter ports org.languagetool.rules.ca.CatalanNumberSpellerFilter.
// SpellNumber converts a digit string (optionally prefixed with "feminine ") to words.
type CatalanNumberSpellerFilter struct {
	// SpellNumber is required for suggestions; nil suppresses all matches.
	SpellNumber func(strToSpell string) string
}

func NewCatalanNumberSpellerFilter(spell func(string) string) *CatalanNumberSpellerFilter {
	return &CatalanNumberSpellerFilter{SpellNumber: spell}
}

// Suggest returns the spelled number, or "" to suppress.
// sentenceStart capitalises the first letter; gender "feminine" prefixes the request.
// wordCountMax is 4 (Java: split length < 4).
func (f *CatalanNumberSpellerFilter) Suggest(numberToSpell, gender string, sentenceStart bool) string {
	if f.SpellNumber == nil {
		return ""
	}
	str := strings.ReplaceAll(numberToSpell, ".", "")
	if gender == "feminine" {
		str = "feminine " + str
	}
	spelled := f.SpellNumber(str)
	if sentenceStart {
		spelled = tools.UppercaseFirstChar(spelled)
	}
	if spelled == "" {
		return ""
	}
	// count words after normalising hyphenated CA forms
	norm := strings.ReplaceAll(spelled, "-i-", " ")
	norm = strings.ReplaceAll(norm, "-", " ")
	parts := strings.Fields(norm)
	if len(parts) >= 4 {
		return ""
	}
	return spelled
}
