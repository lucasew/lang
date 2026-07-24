package fr

import (
	"regexp"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// MakeContractionsFilter ports org.languagetool.rules.fr.MakeContractionsFilter.
type MakeContractionsFilter struct {
	*rules.MakeContractionsFilter
}

// Go RE2 \b is ASCII-only; use Unicode letter boundaries for à/de.
var (
	frDeLe  = regexp.MustCompile(`(?i)(?:^|[^\p{L}])de le(?:[^\p{L}]|$)`)
	frALe   = regexp.MustCompile(`(?i)(?:^|[^\p{L}])à le(?:[^\p{L}]|$)`)
	frDeLes = regexp.MustCompile(`(?i)(?:^|[^\p{L}])de les(?:[^\p{L}]|$)`)
	frALes  = regexp.MustCompile(`(?i)(?:^|[^\p{L}])à les(?:[^\p{L}]|$)`)
)

func NewMakeContractionsFilter() *MakeContractionsFilter {
	return &MakeContractionsFilter{
		MakeContractionsFilter: rules.NewMakeContractionsFilter(fixFrenchContractions),
	}
}

func fixFrenchContractions(suggestion string) string {
	suggestion = replacePhrase(suggestion, frDeLe, "du")
	suggestion = replacePhrase(suggestion, frALe, "au")
	suggestion = replacePhrase(suggestion, frDeLes, "des")
	suggestion = replacePhrase(suggestion, frALes, "aux")
	return suggestion
}

// replacePhrase rewrites "de le"/"à le"/… while preserving surrounding boundary chars.
func replacePhrase(s string, re *regexp.Regexp, repl string) string {
	return re.ReplaceAllStringFunc(s, func(m string) string {
		runes := []rune(m)
		if len(runes) == 0 {
			return m
		}
		lead, trail := "", ""
		start, end := 0, len(runes)
		if !unicode.IsLetter(runes[0]) {
			lead = string(runes[0])
			start = 1
		}
		if end > start && !unicode.IsLetter(runes[end-1]) {
			trail = string(runes[end-1])
			end--
		}
		_ = start
		return lead + repl + trail
	})
}
