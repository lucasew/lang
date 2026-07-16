package de

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// StyleTooOftenUsedVerbRule surface stand-in: lowercase letter tokens len≥4 not stopwords.
type StyleTooOftenUsedVerbRule struct {
	*rules.AbstractStyleTooOftenUsedWordRule
}

func NewStyleTooOftenUsedVerbRule(messages map[string]string) *StyleTooOftenUsedVerbRule {
	base := &rules.AbstractStyleTooOftenUsedWordRule{
		Messages:    messages,
		ID:          "TOO_OFTEN_USED_VERB_DE",
		Description: "Statistische Stilanalyse: Zu häufig genutztes Verb",
		MinPercent:  0,
		MinWords:    0,
		IsCounted: func(tok *languagetool.AnalyzedTokenReadings, index int, tokens []*languagetool.AnalyzedTokenReadings) bool {
			w := tok.GetToken()
			if utf8.RuneCountInString(w) < 4 {
				return false
			}
			// verbs are typically lowercase mid-sentence
			r0, _ := utf8.DecodeRuneInString(w)
			if !unicode.IsLower(r0) {
				return false
			}
			for _, r := range w {
				if !unicode.IsLetter(r) {
					return false
				}
			}
			if _, stop := styleRepeatStop[strings.ToLower(w)]; stop {
				return false
			}
			return true
		},
		Key: func(tok *languagetool.AnalyzedTokenReadings) string {
			return strings.ToLower(tok.GetToken())
		},
		LimitMessage: func(limit int) string {
			return "Das Verb wird mehrfach verwendet. Möglicherweise ist es besser es durch ein Synonym zu ersetzen."
		},
	}
	return &StyleTooOftenUsedVerbRule{AbstractStyleTooOftenUsedWordRule: base}
}

func (r *StyleTooOftenUsedVerbRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractStyleTooOftenUsedWordRule.MatchList(sentences)
}
