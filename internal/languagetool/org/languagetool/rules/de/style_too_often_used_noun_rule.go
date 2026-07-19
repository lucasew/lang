package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// StyleTooOftenUsedNounRule ports org.languagetool.rules.de.StyleTooOftenUsedNounRule.
// Java: SUB: only (no surface invent). Default MinPercent 5, MinWordCount 100.
// Twin tests use MinPercent 0 / MinWordCount 0.
type StyleTooOftenUsedNounRule struct {
	*rules.AbstractStyleTooOftenUsedWordRule
}

const styleTooOftenDefaultMinPercent = 5

func NewStyleTooOftenUsedNounRule(messages map[string]string) *StyleTooOftenUsedNounRule {
	base := &rules.AbstractStyleTooOftenUsedWordRule{
		Messages:     messages,
		ID:           "TOO_OFTEN_USED_NOUN_DE",
		Description:  "Statistische Stilanalyse: Zu häufig genutztes Substantiv",
		MinPercent:   0, // show-all for twin tests
		MinWordCount: 0,
		LimitMessage: func(limit int) string {
			return "Das Substantiv wird häufiger verwendet als " + itoaDE(limit) +
				"% aller Substantive. Möglicherweise ist es besser es durch ein Synonym zu ersetzen."
		},
	}
	base.IsToCountedWord = func(tok *languagetool.AnalyzedTokenReadings) bool {
		return tok != nil && tok.HasPosTagStartingWith("SUB:")
	}
	base.IsException = func(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
		token := tokens[n]
		if token.HasPosTagStartingWith("PRO:") {
			return true
		}
		switch token.GetToken() {
		case "Ich", "Aber", "Ja":
			return true
		}
		// Frau/Herr + EIG or isPosTagUnknown
		if n < len(tokens)-1 && tokens[n+1] != nil &&
			(token.GetToken() == "Frau" || token.GetToken() == "Herr") &&
			(tokens[n+1].HasPosTagStartingWith("EIG:") || !tokens[n+1].IsTagged()) {
			return true
		}
		return false
	}
	base.ToAddedLemma = func(tok *languagetool.AnalyzedTokenReadings) string {
		return rules.LemmaForPosTagStartsWith("SUB:", tok)
	}
	rules.InitStyleTooOftenUsedWordMeta(base, messages, false)
	return &StyleTooOftenUsedNounRule{AbstractStyleTooOftenUsedWordRule: base}
}

func NewStyleTooOftenUsedNounRuleWithDefaultLimit(messages map[string]string) *StyleTooOftenUsedNounRule {
	r := NewStyleTooOftenUsedNounRule(messages)
	r.MinPercent = styleTooOftenDefaultMinPercent
	r.MinWordCount = 100
	return r
}

func (r *StyleTooOftenUsedNounRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractStyleTooOftenUsedWordRule.MatchList(sentences)
}
