package de

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// StyleTooOftenUsedAdjectiveRule surface stand-in: lowercase letter tokens ending in common adj endings.
type StyleTooOftenUsedAdjectiveRule struct {
	*rules.AbstractStyleTooOftenUsedWordRule
}

func NewStyleTooOftenUsedAdjectiveRule(messages map[string]string) *StyleTooOftenUsedAdjectiveRule {
	base := &rules.AbstractStyleTooOftenUsedWordRule{
		Messages:    messages,
		ID:          "TOO_OFTEN_USED_ADJECTIVE_DE",
		Description: "Statistische Stilanalyse: Zu häufig genutztes Adjektiv",
		MinPercent:  0,
		MinWords:    0,
		IsCounted: func(tok *languagetool.AnalyzedTokenReadings, index int, tokens []*languagetool.AnalyzedTokenReadings) bool {
			w := tok.GetToken()
			lc := strings.ToLower(w)
			if utf8.RuneCountInString(lc) < 4 {
				return false
			}
			r0, _ := utf8.DecodeRuneInString(lc)
			if !unicode.IsLower(r0) {
				return false
			}
			for _, r := range lc {
				if !unicode.IsLetter(r) {
					return false
				}
			}
			// weak adj ending heuristic
			for _, suf := range []string{"lich", "isch", "ig", "bar", "sam", "en", "er", "es", "em", "e"} {
				if strings.HasSuffix(lc, suf) && utf8.RuneCountInString(lc) > utf8.RuneCountInString(suf)+2 {
					return true
				}
			}
			return false
		},
		Key: func(tok *languagetool.AnalyzedTokenReadings) string {
			return strings.ToLower(tok.GetToken())
		},
		LimitMessage: func(limit int) string {
			return "Das Adjektiv wird mehrfach verwendet. Möglicherweise ist es besser es durch ein Synonym zu ersetzen."
		},
	}
	return &StyleTooOftenUsedAdjectiveRule{AbstractStyleTooOftenUsedWordRule: base}
}

func (r *StyleTooOftenUsedAdjectiveRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractStyleTooOftenUsedWordRule.MatchList(sentences)
}
