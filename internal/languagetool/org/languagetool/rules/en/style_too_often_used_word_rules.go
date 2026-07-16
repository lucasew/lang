package en

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

var enStyleStop = map[string]struct{}{
	"the": {}, "a": {}, "an": {}, "and": {}, "or": {}, "but": {}, "if": {}, "of": {},
	"to": {}, "in": {}, "on": {}, "at": {}, "for": {}, "from": {}, "with": {}, "by": {},
	"is": {}, "are": {}, "was": {}, "were": {}, "be": {}, "been": {}, "being": {},
	"have": {}, "has": {}, "had": {}, "do": {}, "does": {}, "did": {}, "will": {},
	"would": {}, "could": {}, "should": {}, "may": {}, "might": {}, "must": {},
	"this": {}, "that": {}, "these": {}, "those": {}, "it": {}, "its": {}, "they": {},
	"them": {}, "their": {}, "we": {}, "our": {}, "you": {}, "your": {}, "he": {},
	"she": {}, "his": {}, "her": {}, "not": {}, "no": {}, "as": {}, "so": {}, "than": {},
}

func enLettersOnly(w string) bool {
	if utf8.RuneCountInString(w) < 4 {
		return false
	}
	for _, r := range w {
		if !unicode.IsLetter(r) && r != '-' {
			return false
		}
	}
	return true
}

// StyleTooOftenUsedNounRule surface stand-in: content words length ≥5.
type StyleTooOftenUsedNounRule struct {
	*rules.AbstractStyleTooOftenUsedWordRule
}

func NewStyleTooOftenUsedNounRule(messages map[string]string) *StyleTooOftenUsedNounRule {
	base := &rules.AbstractStyleTooOftenUsedWordRule{
		Messages:    messages,
		ID:          "TOO_OFTEN_USED_NOUN_EN",
		Description: "Style: noun used too often",
		MinPercent:  0,
		MinWords:    0,
		IsCounted: func(tok *languagetool.AnalyzedTokenReadings, index int, tokens []*languagetool.AnalyzedTokenReadings) bool {
			w := tok.GetToken()
			if !enLettersOnly(w) || utf8.RuneCountInString(w) < 5 {
				return false
			}
			lc := strings.ToLower(w)
			if _, stop := enStyleStop[lc]; stop {
				return false
			}
			// prefer noun-ish: not -ly, not common verb endings only
			if strings.HasSuffix(lc, "ly") {
				return false
			}
			return true
		},
		Key: func(tok *languagetool.AnalyzedTokenReadings) string {
			return strings.ToLower(tok.GetToken())
		},
		LimitMessage: func(limit int) string {
			return "This noun is used repeatedly. Consider a synonym."
		},
	}
	return &StyleTooOftenUsedNounRule{AbstractStyleTooOftenUsedWordRule: base}
}

func (r *StyleTooOftenUsedNounRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractStyleTooOftenUsedWordRule.MatchList(sentences)
}

// StyleTooOftenUsedVerbRule surface stand-in: lowercase content words.
type StyleTooOftenUsedVerbRule struct {
	*rules.AbstractStyleTooOftenUsedWordRule
}

func NewStyleTooOftenUsedVerbRule(messages map[string]string) *StyleTooOftenUsedVerbRule {
	base := &rules.AbstractStyleTooOftenUsedWordRule{
		Messages:    messages,
		ID:          "TOO_OFTEN_USED_VERB_EN",
		Description: "Style: verb used too often",
		MinPercent:  0,
		MinWords:    0,
		IsCounted: func(tok *languagetool.AnalyzedTokenReadings, index int, tokens []*languagetool.AnalyzedTokenReadings) bool {
			w := tok.GetToken()
			if !enLettersOnly(w) {
				return false
			}
			r0, _ := utf8.DecodeRuneInString(w)
			if !unicode.IsLower(r0) {
				return false
			}
			lc := strings.ToLower(w)
			if _, stop := enStyleStop[lc]; stop {
				return false
			}
			return true
		},
		Key: func(tok *languagetool.AnalyzedTokenReadings) string {
			return strings.ToLower(tok.GetToken())
		},
		LimitMessage: func(limit int) string {
			return "This verb is used repeatedly. Consider a synonym."
		},
	}
	return &StyleTooOftenUsedVerbRule{AbstractStyleTooOftenUsedWordRule: base}
}

func (r *StyleTooOftenUsedVerbRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractStyleTooOftenUsedWordRule.MatchList(sentences)
}

// StyleTooOftenUsedAdjectiveRule surface stand-in: -ly/-ful/-ous/-ive/-able endings.
type StyleTooOftenUsedAdjectiveRule struct {
	*rules.AbstractStyleTooOftenUsedWordRule
}

func NewStyleTooOftenUsedAdjectiveRule(messages map[string]string) *StyleTooOftenUsedAdjectiveRule {
	base := &rules.AbstractStyleTooOftenUsedWordRule{
		Messages:    messages,
		ID:          "TOO_OFTEN_USED_ADJECTIVE_EN",
		Description: "Style: adjective used too often",
		MinPercent:  0,
		MinWords:    0,
		IsCounted: func(tok *languagetool.AnalyzedTokenReadings, index int, tokens []*languagetool.AnalyzedTokenReadings) bool {
			w := tok.GetToken()
			if !enLettersOnly(w) {
				return false
			}
			lc := strings.ToLower(w)
			for _, suf := range []string{"ful", "ous", "ive", "able", "ible", "al", "ic", "ish", "less", "y"} {
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
			return "This adjective is used repeatedly. Consider a synonym."
		},
	}
	return &StyleTooOftenUsedAdjectiveRule{AbstractStyleTooOftenUsedWordRule: base}
}

func (r *StyleTooOftenUsedAdjectiveRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractStyleTooOftenUsedWordRule.MatchList(sentences)
}
