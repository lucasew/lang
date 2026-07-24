package en

import (
	"fmt"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

const defaultStyleMinPercent = 5

// NewStyleTooOftenUsedNounRule ports org.languagetool.rules.en.StyleTooOftenUsedNounRule.
func NewStyleTooOftenUsedNounRule() *rules.AbstractStyleTooOftenUsedWordRule {
	base := &rules.AbstractStyleTooOftenUsedWordRule{
		ID:          "TOO_OFTEN_USED_NOUN_EN",
		Description: "Statistical Style Analysis: Overused Noun",
		MinPercent:  defaultStyleMinPercent,
		// Java: isToCountedWord = NN*; isException = NNP|IN|JJ|RB|VB
		IsToCountedWord: func(tok *languagetool.AnalyzedTokenReadings) bool {
			return tok != nil && tok.HasPosTagStartingWith("NN")
		},
		IsException: func(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
			if n < 0 || n >= len(tokens) || tokens[n] == nil {
				return false
			}
			tok := tokens[n]
			return tok.HasPosTagStartingWith("NNP") || tok.HasPosTagStartingWith("IN") ||
				tok.HasPosTagStartingWith("JJ") || tok.HasPosTagStartingWith("RB") ||
				tok.HasPosTagStartingWith("VB")
		},
		ToAddedLemma: func(tok *languagetool.AnalyzedTokenReadings) string {
			return lemmaForPosPrefix(tok, "NN")
		},
		LimitMessage: func(limit int) string {
			return fmt.Sprintf("The noun is used more than %d%% times of all nouns. It may be better to replace it with a synonym.", limit)
		},
	}
	rules.InitStyleTooOftenUsedWordMeta(base, nil, false)
	return base
}

// NewStyleTooOftenUsedVerbRule ports org.languagetool.rules.en.StyleTooOftenUsedVerbRule.
func NewStyleTooOftenUsedVerbRule() *rules.AbstractStyleTooOftenUsedWordRule {
	base := &rules.AbstractStyleTooOftenUsedWordRule{
		ID:          "TOO_OFTEN_USED_VERB_EN",
		Description: "Statistical Style Analysis: Overused Verb",
		MinPercent:  defaultStyleMinPercent,
		// Java: isToCountedWord = VB*; isException = be|have|do lemmas | IN | NN
		IsToCountedWord: func(tok *languagetool.AnalyzedTokenReadings) bool {
			return tok != nil && tok.HasPosTagStartingWith("VB")
		},
		IsException: func(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
			if n < 0 || n >= len(tokens) || tokens[n] == nil {
				return false
			}
			tok := tokens[n]
			return tok.HasAnyLemma("be", "have", "do") || tok.HasPosTagStartingWith("IN") ||
				tok.HasPosTagStartingWith("NN")
		},
		ToAddedLemma: func(tok *languagetool.AnalyzedTokenReadings) string {
			return lemmaForPosPrefix(tok, "VB")
		},
		LimitMessage: func(limit int) string {
			return fmt.Sprintf("The verb is used more than %d%% times of all verbs. It may be better to replace it with a synonym.", limit)
		},
	}
	rules.InitStyleTooOftenUsedWordMeta(base, nil, false)
	return base
}

// NewStyleTooOftenUsedAdjectiveRule ports org.languagetool.rules.en.StyleTooOftenUsedAdjectiveRule.
func NewStyleTooOftenUsedAdjectiveRule() *rules.AbstractStyleTooOftenUsedWordRule {
	base := &rules.AbstractStyleTooOftenUsedWordRule{
		ID:          "TOO_OFTEN_USED_ADJECTIVE_EN",
		Description: "Statistical Style Analysis: Overused Adjective",
		MinPercent:  defaultStyleMinPercent,
		// Java: isToCountedWord = JJ*; isException = RB|IN|CD|DT|NN
		IsToCountedWord: func(tok *languagetool.AnalyzedTokenReadings) bool {
			return tok != nil && tok.HasPosTagStartingWith("JJ")
		},
		IsException: func(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
			if n < 0 || n >= len(tokens) || tokens[n] == nil {
				return false
			}
			tok := tokens[n]
			return tok.HasPosTagStartingWith("RB") || tok.HasPosTagStartingWith("IN") ||
				tok.HasPosTagStartingWith("CD") || tok.HasPosTagStartingWith("DT") ||
				tok.HasPosTagStartingWith("NN")
		},
		ToAddedLemma: func(tok *languagetool.AnalyzedTokenReadings) string {
			return lemmaForPosPrefix(tok, "JJ")
		},
		LimitMessage: func(limit int) string {
			return fmt.Sprintf("The adjective is used more than %d%% times of all adjectives. It may be better to replace it with a synonym.", limit)
		},
	}
	rules.InitStyleTooOftenUsedWordMeta(base, nil, false)
	return base
}

func lemmaForPosPrefix(tok *languagetool.AnalyzedTokenReadings, prefix string) string {
	for _, r := range tok.GetReadings() {
		if r.GetPOSTag() != nil && strings.HasPrefix(*r.GetPOSTag(), prefix) {
			if r.GetLemma() != nil && *r.GetLemma() != "" {
				return *r.GetLemma()
			}
		}
	}
	return strings.ToLower(tok.GetToken())
}
