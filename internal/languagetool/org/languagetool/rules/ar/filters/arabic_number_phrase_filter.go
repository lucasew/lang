package filters

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ArabicNumberPhraseFilter ports suggestion helpers from
// org.languagetool.rules.ar.filters.ArabicNumberPhraseFilter.
type ArabicNumberPhraseFilter struct{}

func NewArabicNumberPhraseFilter() *ArabicNumberPhraseFilter {
	return &ArabicNumberPhraseFilter{}
}

// PrepareSuggestion builds suggestions for a numeric phrase (digit or known).
// previousWord is optionally prefixed.
func PrepareSuggestion(numPhrase, previousWord string, feminine bool) []string {
	sug := SuggestionsForNumericPhrase(numPhrase, feminine)
	if len(sug) == 0 {
		return nil
	}
	out := make([]string, 0, len(sug))
	for _, s := range sug {
		if previousWord != "" {
			out = append(out, previousWord+" "+s)
		} else {
			out = append(out, s)
		}
	}
	return out
}

// PrepareSuggestionWithUnit appends a unit form after the numeric phrase.
func PrepareSuggestionWithUnit(numPhrase, previousWord, unit, inflection string, feminine bool) []string {
	base := PrepareSuggestion(numPhrase, previousWord, feminine)
	if unit == "" {
		return base
	}
	unitForm := tools.GetArabicUnitOneForm(unit, inflection)
	if inflection == "" {
		unitForm = tools.GetArabicUnitOneForm(unit, "raf3")
	}
	// rough agreement: use plural for numbers >= 3
	if n, err := parseLeadingInt(numPhrase); err == nil {
		if n == 2 {
			unitForm = tools.GetArabicUnitTwoForm(unit, orDefault(inflection, "raf3"))
		} else if n >= 3 && n <= 10 {
			unitForm = tools.GetArabicUnitPluralForm(unit, orDefault(inflection, "raf3"))
		}
	}
	out := make([]string, 0, len(base))
	for _, s := range base {
		out = append(out, s+" "+unitForm)
	}
	return out
}

// SuggestionsForNumericPhrase converts a phrase of digits to Arabic words.
func SuggestionsForNumericPhrase(numPhrase string, feminine bool) []string {
	numPhrase = strings.TrimSpace(numPhrase)
	if numPhrase == "" {
		return nil
	}
	// pure integer
	if isAllDigits(numPhrase) {
		w := tools.NumberToArabicWordsGender(numPhrase, feminine)
		if w == "" {
			return nil
		}
		return []string{w}
	}
	// extract first integer token
	for _, tok := range strings.Fields(numPhrase) {
		if isAllDigits(tok) {
			w := tools.NumberToArabicWordsGender(tok, feminine)
			if w != "" {
				return []string{w}
			}
		}
	}
	return nil
}

// InflectionFromPrevious returns "jar" when previous token starts with ب/ل/ك.
func InflectionFromPrevious(previousWord string) string {
	if previousWord == "" {
		return ""
	}
	r := []rune(previousWord)
	if len(r) == 0 {
		return ""
	}
	switch r[0] {
	case 'ب', 'ل', 'ك':
		return "jar"
	}
	return ""
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func parseLeadingInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if b.Len() > 0 {
			break
		}
	}
	return strconv.Atoi(b.String())
}

func orDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}
