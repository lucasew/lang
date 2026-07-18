package tokenizers

import (
	"strings"
	"unicode"
)

// SimpleSentenceTokenizer ports Default rules from segment-simple.srx:
// break after [.!?…] followed by whitespace, or [.!?…] followed by uppercase.
// Soft: also applies segment.srx-style no-break after common abbreviations (etc.).
type SimpleSentenceTokenizer struct{}

func NewSimpleSentenceTokenizer() *SimpleSentenceTokenizer {
	return &SimpleSentenceTokenizer{}
}

// Common abbreviations that must not end a sentence when followed by space
// (subset of LanguageTool segment.srx beforebreak rules, multi-lang soft path).
var noBreakAbbrevs = map[string]struct{}{
	"etc": {}, "șamd": {}, "samd": {}, "vs": {}, "cf": {}, "al": {},
	"mr": {}, "mrs": {}, "ms": {}, "dr": {}, "prof": {}, "sr": {}, "jr": {},
	"fig": {}, "vol": {}, "pp": {}, "no": {}, "st": {}, "inc": {}, "ltd": {},
	"corp": {}, "co": {}, "approx": {}, "e.g": {}, "i.e": {}, "ex": {},
	"art": {}, "cap": {}, "tel": {}, "op": {}, "esp": {},
}

// Tokenize returns sentence segments that concatenate back to text.
func (t *SimpleSentenceTokenizer) Tokenize(text string) []string {
	if text == "" {
		return nil
	}
	var out []string
	start := 0
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r != '.' && r != '!' && r != '?' && r != '…' {
			continue
		}
		// consume run of sentence-ending punctuation
		j := i
		for j+1 < len(runes) {
			n := runes[j+1]
			if n == '.' || n == '!' || n == '?' || n == '…' {
				j++
				continue
			}
			break
		}
		// Soft SRX: do not break after abbreviation periods (etc. etc.).
		if r == '.' && isAbbrevPeriod(runes, i) {
			i = j
			continue
		}
		// Soft: do not break after ordinal/day numbers (18. Main, 12. bis 11. Januar).
		// German dates and ordinals use digit+period without ending the sentence.
		if r == '.' && isOrdinalNumberPeriod(runes, i) {
			i = j
			continue
		}
		// case 1: punct + whitespace → break after one whitespace
		if j+1 < len(runes) && unicode.IsSpace(runes[j+1]) {
			end := j + 2 // include one whitespace (SRX \s)
			if end > len(runes) {
				end = len(runes)
			}
			out = append(out, string(runes[start:end]))
			start = end
			i = end - 1
			continue
		}
		// case 2: punct + uppercase → break before uppercase
		if j+1 < len(runes) && unicode.IsUpper(runes[j+1]) {
			end := j + 1
			out = append(out, string(runes[start:end]))
			start = end
			i = end - 1
			continue
		}
		i = j
	}
	if start < len(runes) {
		out = append(out, string(runes[start:]))
	}
	return out
}

// isOrdinalNumberPeriod is true when the period follows one or more digits
// (German ordinal/day: "18. Mai", "12. bis 11. Januar").
func isOrdinalNumberPeriod(runes []rune, dotIdx int) bool {
	if dotIdx <= 0 || runes[dotIdx] != '.' {
		return false
	}
	j := dotIdx - 1
	if j < 0 || !unicode.IsDigit(runes[j]) {
		return false
	}
	// consume full digit run (optional)
	for j > 0 && unicode.IsDigit(runes[j-1]) {
		j--
	}
	return true
}

// isAbbrevPeriod reports whether runes[dotIdx] is a period after a known abbreviation.
func isAbbrevPeriod(runes []rune, dotIdx int) bool {
	if dotIdx <= 0 || runes[dotIdx] != '.' {
		return false
	}
	// Collect word characters immediately before the period.
	end := dotIdx
	start := end
	for start > 0 {
		r := runes[start-1]
		if unicode.IsLetter(r) || r == '\'' || r == '’' {
			start--
			continue
		}
		// allow internal dots in e.g. / i.e. (period already consumed as end)
		break
	}
	if start == end {
		return false
	}
	word := strings.ToLower(string(runes[start:end]))
	// multi-dot abbrevs written as e.g. or i.e. — take last segment after last letter run
	if _, ok := noBreakAbbrevs[word]; ok {
		return true
	}
	return false
}
