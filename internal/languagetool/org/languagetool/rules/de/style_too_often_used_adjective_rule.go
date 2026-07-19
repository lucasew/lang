package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// StyleTooOftenUsedAdjectiveRule ports org.languagetool.rules.de.StyleTooOftenUsedAdjectiveRule.
// Java: ADJ: only — no surface invent.
type StyleTooOftenUsedAdjectiveRule struct {
	*rules.AbstractStyleTooOftenUsedWordRule
}

func NewStyleTooOftenUsedAdjectiveRule(messages map[string]string) *StyleTooOftenUsedAdjectiveRule {
	base := &rules.AbstractStyleTooOftenUsedWordRule{
		Messages:     messages,
		ID:           "TOO_OFTEN_USED_ADJECTIVE_DE",
		Description:  "Statistische Stilanalyse: Zu häufig genutztes Adjektiv",
		MinPercent:   0,
		MinWordCount: 0,
		LimitMessage: func(limit int) string {
			// Java has a space before the period: "Adjektive . "
			return "Das Adjektiv wird häufiger verwendet als " + itoaDE(limit) +
				"% aller Adjektive . Möglicherweise ist es besser es durch ein Synonym zu ersetzen."
		},
	}
	base.IsToCountedWord = func(tok *languagetool.AnalyzedTokenReadings) bool {
		return tok != nil && tok.HasPosTagStartingWith("ADJ:")
	}
	base.IsException = func(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
		token := tokens[n]
		return token.HasPosTagStartingWith("PRO:") ||
			token.HasPosTagStartingWith("ADV:") ||
			token.HasPosTagStartingWith("ZUS")
	}
	base.ToAddedLemma = func(tok *languagetool.AnalyzedTokenReadings) string {
		return rules.LemmaForPosTagStartsWith("ADJ:", tok)
	}
	rules.InitStyleTooOftenUsedWordMeta(base, messages, false)
	return &StyleTooOftenUsedAdjectiveRule{AbstractStyleTooOftenUsedWordRule: base}
}

func NewStyleTooOftenUsedAdjectiveRuleWithDefaultLimit(messages map[string]string) *StyleTooOftenUsedAdjectiveRule {
	r := NewStyleTooOftenUsedAdjectiveRule(messages)
	r.MinPercent = styleTooOftenDefaultMinPercent
	r.MinWordCount = 100
	return r
}

func (r *StyleTooOftenUsedAdjectiveRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractStyleTooOftenUsedWordRule.MatchList(sentences)
}
