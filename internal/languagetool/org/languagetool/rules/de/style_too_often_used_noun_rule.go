package de

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// StyleTooOftenUsedNounRule is a surface stand-in for StyleTooOftenUsedNounRule.
// Without SUB: POS tags, mid-sentence capitalized tokens (German noun heuristic) are counted.
type StyleTooOftenUsedNounRule struct {
	*rules.AbstractStyleTooOftenUsedWordRule
}

func NewStyleTooOftenUsedNounRule(messages map[string]string) *StyleTooOftenUsedNounRule {
	base := &rules.AbstractStyleTooOftenUsedWordRule{
		Messages:    messages,
		ID:          "TOO_OFTEN_USED_NOUN_DE",
		Description: "Statistische Stilanalyse: Zu häufig genutztes Substantiv",
		MinPercent:  0, // show-all for surface tests
		MinWords:    0,
		IsCounted: func(tok *languagetool.AnalyzedTokenReadings, index int, tokens []*languagetool.AnalyzedTokenReadings) bool {
			w := tok.GetToken()
			if utf8.RuneCountInString(w) < 3 {
				return false
			}
			// sentence-start capital is ambiguous; skip first content token
			if index <= 1 {
				return false
			}
			if !tools.StartsWithUppercase(w) {
				return false
			}
			for _, r := range w {
				if !unicode.IsLetter(r) && r != '-' {
					return false
				}
			}
			// skip some interjections
			switch w {
			case "Ich", "Aber", "Ja", "Nein", "Doch":
				return false
			}
			return true
		},
		Key: func(tok *languagetool.AnalyzedTokenReadings) string {
			return strings.ToLower(tok.GetToken())
		},
		LimitMessage: func(limit int) string {
			if limit == 0 {
				return "Das Substantiv wird mehrfach verwendet. Möglicherweise ist es besser es durch ein Synonym zu ersetzen."
			}
			return "Das Substantiv wird häufiger verwendet als " + itoa(limit) + "% aller Substantive. Möglicherweise ist es besser es durch ein Synonym zu ersetzen."
		},
	}
	return &StyleTooOftenUsedNounRule{AbstractStyleTooOftenUsedWordRule: base}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func (r *StyleTooOftenUsedNounRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractStyleTooOftenUsedWordRule.MatchList(sentences)
}
