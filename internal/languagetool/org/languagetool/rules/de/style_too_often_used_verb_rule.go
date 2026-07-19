package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// StyleTooOftenUsedVerbRule ports org.languagetool.rules.de.StyleTooOftenUsedVerbRule.
// Java: VER: only — no surface invent.
type StyleTooOftenUsedVerbRule struct {
	*rules.AbstractStyleTooOftenUsedWordRule
}

func NewStyleTooOftenUsedVerbRule(messages map[string]string) *StyleTooOftenUsedVerbRule {
	base := &rules.AbstractStyleTooOftenUsedWordRule{
		Messages:     messages,
		ID:           "TOO_OFTEN_USED_VERB_DE",
		Description:  "Statistische Stilanalyse: Zu häufig genutztes Verb",
		MinPercent:   0,
		MinWordCount: 0,
		LimitMessage: func(limit int) string {
			return "Das Verb wird häufiger verwendet als " + itoaDE(limit) +
				"% aller Verben. Möglicherweise ist es besser es durch ein Synonym zu ersetzen."
		},
	}
	base.IsToCountedWord = func(tok *languagetool.AnalyzedTokenReadings) bool {
		return tok != nil && tok.HasPosTagStartingWith("VER:")
	}
	base.IsException = func(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
		token := tokens[n]
		return token.HasPosTagStartingWith("VER:MOD") ||
			token.HasPosTagStartingWith("VER:AUX") ||
			token.HasPosTagStartingWith("ART") ||
			token.HasPosTagStartingWith("ADJ")
	}
	base.ToAddedLemma = func(tok *languagetool.AnalyzedTokenReadings) string {
		return rules.LemmaForPosTagStartsWith("VER:", tok)
	}
	rules.InitStyleTooOftenUsedWordMeta(base, messages, false)
	return &StyleTooOftenUsedVerbRule{AbstractStyleTooOftenUsedWordRule: base}
}

func NewStyleTooOftenUsedVerbRuleWithDefaultLimit(messages map[string]string) *StyleTooOftenUsedVerbRule {
	r := NewStyleTooOftenUsedVerbRule(messages)
	r.MinPercent = styleTooOftenDefaultMinPercent
	r.MinWordCount = 100
	return r
}

func (r *StyleTooOftenUsedVerbRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractStyleTooOftenUsedWordRule.MatchList(sentences)
}
