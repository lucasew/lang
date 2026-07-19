package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SentenceWithManRule ports org.languagetool.rules.de.SentenceWithManRule
// (AbstractStatisticSentenceStyleRule; default off; DEFAULT_MIN_PERCENT=15; ‰).
type SentenceWithManRule struct {
	*rules.AbstractStatisticSentenceStyleRule
}

const sentenceWithManDefaultMinPercent = 15

func NewSentenceWithManRule(messages map[string]string) *SentenceWithManRule {
	r := &SentenceWithManRule{
		AbstractStatisticSentenceStyleRule: &rules.AbstractStatisticSentenceStyleRule{
			ID:                  "SENTENCE_WITH_MAN_DE",
			Description:         "Statistische Stilanalyse: Sätze mit indirekter Leseransprache 'man'",
			MinPercent:          0, // twin tests / Match show all; Java default 15‰
			ExcludeDirectSpeech: true,
			Denominator:         1000,
		},
	}
	r.ConditionFulfilled = r.conditionFulfilled
	r.LimitMessage = r.getLimitMessage
	rules.InitStatisticSentenceStyleMeta(r.AbstractStatisticSentenceStyleRule, messages, false)
	return r
}

func NewSentenceWithManRuleWithDefaultLimit(messages map[string]string) *SentenceWithManRule {
	r := NewSentenceWithManRule(messages)
	r.MinPercent = sentenceWithManDefaultMinPercent
	return r
}

func (r *SentenceWithManRule) GetID() string {
	if r != nil && r.AbstractStatisticSentenceStyleRule != nil {
		return r.AbstractStatisticSentenceStyleRule.GetID()
	}
	return "SENTENCE_WITH_MAN_DE"
}

func (r *SentenceWithManRule) GetDescription() string {
	return "Statistische Stilanalyse: Sätze mit indirekter Leseransprache 'man'"
}

func (r *SentenceWithManRule) getLimitMessage(limit int, percent float64) string {
	if limit == 0 {
		return "Sätze mit der indirekten Leseransprache 'man' sind stilistisch wenig elegant formuliert. Lässt sich das Wort vermeiden?"
	}
	return "Mehr als " + itoaDE(limit) + "‰ Sätze mit der indirekten Leseransprache 'man' {" +
		itoaDE(int(percent+0.5)) + "‰} gefunden. Lässt sich das Wort vermeiden?"
}

func (r *SentenceWithManRule) conditionFulfilled(sentence []*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedTokenReadings {
	for _, t := range sentence {
		if t == nil {
			continue
		}
		if isWordMan(t) {
			return t
		}
	}
	return nil
}

func isWordMan(token *languagetool.AnalyzedTokenReadings) bool {
	// Java: token.hasLemma("man") only
	return token != nil && token.HasAnyLemma("man")
}

// Match single-sentence convenience (Java is text-level MatchList).
func (r *SentenceWithManRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil {
		return nil
	}
	return r.MatchList([]*languagetool.AnalyzedSentence{sentence})
}
