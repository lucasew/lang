package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SentenceWithModalVerbRule ports org.languagetool.rules.de.SentenceWithModalVerbRule
// (AbstractStatisticSentenceStyleRule; default off; DEFAULT_MIN_PERCENT=18).
// Java: VER:MOD + VER:INF (either order, mark breaks) — no surface form invent.
type SentenceWithModalVerbRule struct {
	*rules.AbstractStatisticSentenceStyleRule
}

const sentenceWithModalDefaultMinPercent = 18

func NewSentenceWithModalVerbRule(messages map[string]string) *SentenceWithModalVerbRule {
	r := &SentenceWithModalVerbRule{
		AbstractStatisticSentenceStyleRule: &rules.AbstractStatisticSentenceStyleRule{
			ID:                  "SENTENCE_WITH_MODAL_VERB_DE",
			Description:         "Statistische Stilanalyse: Sätze mit Modalverben",
			MinPercent:          sentenceWithModalDefaultMinPercent, // Java DEFAULT_MIN_PERCENT=18
			ExcludeDirectSpeech: true,
			Denominator:         100,
		},
	}
	r.ConditionFulfilled = r.conditionFulfilled
	r.LimitMessage = r.getLimitMessage
	rules.InitStatisticSentenceStyleMeta(r.AbstractStatisticSentenceStyleRule, messages, false)
	return r
}

func NewSentenceWithModalVerbRuleWithDefaultLimit(messages map[string]string) *SentenceWithModalVerbRule {
	return NewSentenceWithModalVerbRule(messages)
}

func NewSentenceWithModalVerbRuleWithMinPercent(messages map[string]string, min int) *SentenceWithModalVerbRule {
	r := NewSentenceWithModalVerbRule(messages)
	r.MinPercent = min
	return r
}

func (r *SentenceWithModalVerbRule) GetID() string {
	if r != nil && r.AbstractStatisticSentenceStyleRule != nil {
		return r.AbstractStatisticSentenceStyleRule.GetID()
	}
	return "SENTENCE_WITH_MODAL_VERB_DE"
}

func (r *SentenceWithModalVerbRule) GetDescription() string {
	return "Statistische Stilanalyse: Sätze mit Modalverben"
}

func (r *SentenceWithModalVerbRule) getLimitMessage(limit int, percent float64) string {
	if limit == 0 {
		return "Modalverb: Modalverben blähen den Text häufig auf und sollten vermieden werden."
	}
	return "Mehr als " + itoaDE(limit) + "% Sätze mit Modalverben {" + itoaDE(int(percent+0.5)) +
		"%} gefunden. Modalverben blähen den Text häufig auf und sollten vermieden werden."
}

func (r *SentenceWithModalVerbRule) isModalVerb(token *languagetool.AnalyzedTokenReadings) bool {
	// Java: hasPosTagStartingWith("VER:MOD") only
	return token != nil && token.HasPosTagStartingWith("VER:MOD")
}

func isInfinitiveModal(token *languagetool.AnalyzedTokenReadings) bool {
	// Java: hasPosTagStartingWith("VER:INF") only
	return token != nil && token.HasPosTagStartingWith("VER:INF")
}

func (r *SentenceWithModalVerbRule) conditionFulfilled(sentence []*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedTokenReadings {
	// Java: modal then INF, or INF then modal; marks abort. No modal-only hits.
	for i := 0; i < len(sentence); i++ {
		if r.isModalVerb(sentence[i]) {
			token := sentence[i]
			for j := i + 1; j < len(sentence); j++ {
				if isInfinitiveModal(sentence[j]) {
					return token
				} else if rules.IsStatisticMark(sentence[j]) {
					return nil
				}
			}
		} else if isInfinitiveModal(sentence[i]) {
			for j := i + 1; j < len(sentence); j++ {
				if r.isModalVerb(sentence[j]) {
					return sentence[j]
				} else if rules.IsStatisticMark(sentence[j]) {
					return nil
				}
			}
		}
	}
	return nil
}

func (r *SentenceWithModalVerbRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil {
		return nil
	}
	return r.MatchList([]*languagetool.AnalyzedSentence{sentence})
}
