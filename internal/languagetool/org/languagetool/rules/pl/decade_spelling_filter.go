package pl

import (
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DecadeSpellingFilter ports org.languagetool.rules.pl.DecadeSpellingFilter.
// Used by DATA_DEKADY grammar rules (args lata:…).
type DecadeSpellingFilter struct{}

func NewDecadeSpellingFilter() *DecadeSpellingFilter {
	return &DecadeSpellingFilter{}
}

// AcceptRuleMatch ports DecadeSpellingFilter.acceptRuleMatch.
// Rewrites message placeholders {dekada} and {wiek} from arguments["lata"].
// Returns nil when lata is missing/unparseable (Java catches IllegalArgumentException).
func (f *DecadeSpellingFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil {
		return nil
	}
	lata := ""
	if arguments != nil {
		lata = arguments["lata"]
	}
	msg := f.FormatMessage(match.GetMessage(), lata)
	if msg == "" {
		return nil
	}
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), msg)
	out.ShortMessage = match.ShortMessage
	return out
}

// FormatMessage replaces {dekada} and {wiek} in the rule message.
// lata: first 2 chars = century base, rest from index 2 = decade (Java substring).
// Returns empty string if lata is unparseable (maps to null match).
func (f *DecadeSpellingFilter) FormatMessage(message, lata string) string {
	// Java: century = substring(0,2), decade = substring(2); parseInt(century).
	// Need at least 2 chars for century; shorter strings throw in Java (fail closed → "").
	if len(lata) < 2 {
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

// toRoman ports DecadeSpellingFilter.getRomanNumber.
func toRoman(num int) string {
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
