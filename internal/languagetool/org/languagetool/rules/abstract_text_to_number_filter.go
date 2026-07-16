package rules

import (
	"math"
	"strconv"
	"strings"
)

// TextToNumberFilter ports org.languagetool.rules.AbstractTextToNumberFilter.
// Language modules supply numbers/multipliers maps and comma/percentage hooks.
type TextToNumberFilter struct {
	Numbers     map[string]float64
	Multipliers map[string]float64
	// IsComma reports decimal separators written as words ("coma", "comma", …).
	IsComma func(s string) bool
	// IsPercentage reports percentage markers at token i (e.g. "por" + "ciento").
	// tokens are the original surface forms of the pattern tokens.
	IsPercentage func(tokens []string, i int) bool
	// Tokenize splits a single token into sub-forms (CA: hyphen). Nil → whole token.
	Tokenize func(s string) []string
	// FormatResult post-processes the numeric string (CA: '.' → ','). Nil → identity.
	FormatResult func(s string) string
}

// Convert parses written-out numbers inside a match span and returns the formatted
// numeric suggestion (Java acceptRuleMatch adds this as a suggested replacement).
//
// tokens are the pattern tokens (surface forms). fromPos/toPos are character
// offsets into the original sentence; tokenStarts/tokenEnds are the start/end
// offsets of each token (same units as fromPos/toPos). When offsets are omitted
// (nil or shorter than tokens), all tokens are treated as inside the match.
func (f *TextToNumberFilter) Convert(tokens []string, fromPos, toPos int, tokenStarts, tokenEnds []int) string {
	numbers := f.Numbers
	if numbers == nil {
		numbers = map[string]float64{}
	}
	multipliers := f.Multipliers
	if multipliers == nil {
		multipliers = map[string]float64{}
	}
	tokenize := f.Tokenize
	if tokenize == nil {
		tokenize = func(s string) []string { return []string{s} }
	}
	isComma := f.IsComma
	if isComma == nil {
		isComma = func(string) bool { return false }
	}
	isPct := f.IsPercentage
	if isPct == nil {
		isPct = func([]string, int) bool { return false }
	}

	var total, current, currentDecimal float64
	var addedZeros int
	percentage := false
	decimal := false

	for posWord := 0; posWord < len(tokens); posWord++ {
		if len(tokenEnds) > posWord && tokenEnds[posWord] > toPos && len(tokenStarts) > 0 {
			break
		}
		inside := true
		if len(tokenStarts) > posWord && len(tokenEnds) > posWord {
			inside = tokenStarts[posWord] >= fromPos && tokenEnds[posWord] <= toPos
		}
		if !inside {
			continue
		}
		form := strings.ToLower(tokens[posWord])
		if posWord > 0 && isPct(tokens, posWord) {
			percentage = true
			break
		}
		if isComma(form) {
			decimal = true
			continue
		}
		for _, subForm := range tokenize(form) {
			sub := strings.ToLower(subForm)
			if !decimal {
				if v, ok := numbers[sub]; ok {
					current += v
				} else if m, ok := multipliers[sub]; ok {
					if current == 0 { // mil
						current = 1
					}
					total += current * m
					current = 0
				}
			} else if v, ok := numbers[sub]; ok {
				zerosToAdd := len(formatNumber(v, false))
				currentDecimal += v / math.Pow(10, float64(addedZeros+zerosToAdd))
				addedZeros++
			}
		}
	}
	total += current
	total += currentDecimal
	return f.format(total, percentage)
}

func (f *TextToNumberFilter) format(d float64, percentage bool) string {
	result := formatNumber(d, percentage)
	if f.FormatResult != nil {
		return f.FormatResult(result)
	}
	return result
}

func formatNumber(d float64, percentage bool) string {
	var result string
	if d == float64(int64(d)) && !math.IsInf(d, 0) && math.Abs(d) < 1e15 {
		result = strconv.FormatInt(int64(d), 10)
	} else {
		result = strconv.FormatFloat(d, 'f', -1, 64)
	}
	if percentage {
		result = result + "\u202F%" // narrow non-breaking space + percentage
	}
	return result
}

// ConvertTokens is a convenience when all tokens lie inside the match.
func (f *TextToNumberFilter) ConvertTokens(tokens []string) string {
	return f.Convert(tokens, 0, 0, nil, nil)
}
