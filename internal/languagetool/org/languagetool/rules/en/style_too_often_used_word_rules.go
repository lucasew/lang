package en

import (
	"fmt"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// defaultStyleMinPercent is Java StyleTooOftenUsed{Noun,Verb,Adjective}Rule.DEFAULT_MIN_PERCENT.
const defaultStyleMinPercent = 5

// NewStyleTooOftenUsedNounRule ports org.languagetool.rules.en.StyleTooOftenUsedNounRule.
func NewStyleTooOftenUsedNounRule() *rules.AbstractStyleTooOftenUsedWordRule {
	base := &rules.AbstractStyleTooOftenUsedWordRule{
		ID:                "TOO_OFTEN_USED_NOUN_EN",
		Description:       "Statistical Style Analysis: Overused Noun",
		ConfigureText:     "Show when a noun is used more often than ...% of all nouns:",
		DefaultMinPercent: defaultStyleMinPercent,
		MinPercent:        defaultStyleMinPercent,
		MinWordCount:      100, // Java AbstractStyleTooOftenUsedWordRule.MIN_WORD_COUNT
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
			// Java: getLemmaForPosTagStartsWith("NN", token) — null when no matching reading.
			return rules.LemmaForPosTagStartsWith("NN", tok)
		},
		LimitMessage: func(limit int) string {
			// Java: "The noun is used more than " + limit + "% times of all nouns. " +
			// "It may be better to replace it with a synonym."
			return fmt.Sprintf("The noun is used more than %d%% times of all nouns. It may be better to replace it with a synonym.", limit)
		},
	}
	rules.InitStyleTooOftenUsedWordMeta(base, nil, false)
	return base
}

// NewStyleTooOftenUsedVerbRule ports org.languagetool.rules.en.StyleTooOftenUsedVerbRule.
func NewStyleTooOftenUsedVerbRule() *rules.AbstractStyleTooOftenUsedWordRule {
	base := &rules.AbstractStyleTooOftenUsedWordRule{
		ID:                "TOO_OFTEN_USED_VERB_EN",
		Description:       "Statistical Style Analysis: Overused Verb",
		ConfigureText:     "Show when a verb is used more often than ...% of all verbs:",
		DefaultMinPercent: defaultStyleMinPercent,
		MinPercent:        defaultStyleMinPercent,
		MinWordCount:      100, // Java AbstractStyleTooOftenUsedWordRule.MIN_WORD_COUNT
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
			// Java: getLemmaForPosTagStartsWith("VB", token)
			return rules.LemmaForPosTagStartsWith("VB", tok)
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
		ID:                "TOO_OFTEN_USED_ADJECTIVE_EN",
		Description:       "Statistical Style Analysis: Overused Adjective",
		// Java bug-for-bug: "a adjective" (not "an").
		ConfigureText:     "Show when a adjective is used more often than ...% of all adjectives:",
		DefaultMinPercent: defaultStyleMinPercent,
		MinPercent:        defaultStyleMinPercent,
		MinWordCount:      100, // Java AbstractStyleTooOftenUsedWordRule.MIN_WORD_COUNT
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
			// Java: getLemmaForPosTagStartsWith("JJ", token)
			return rules.LemmaForPosTagStartsWith("JJ", tok)
		},
		LimitMessage: func(limit int) string {
			// Java: "The adjective is used more than " + limit + "% times of all adjectives. " +
			// "It may be better to replace it with a synonym."
			return fmt.Sprintf("The adjective is used more than %d%% times of all adjectives. It may be better to replace it with a synonym.", limit)
		},
	}
	rules.InitStyleTooOftenUsedWordMeta(base, nil, false)
	return base
}
