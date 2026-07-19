package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PassiveSentenceRule ports org.languagetool.rules.de.PassiveSentenceRule
// (extends AbstractStatisticSentenceStyleRule; default off, DEFAULT_MIN_PERCENT=8).
// Match() uses MinPercent for single-sentence convenience (Java is text-level MatchList).
// Java: hasLemma("werden") + hasPosTagStartingWith("VER:PA2:") — no surface invent.
type PassiveSentenceRule struct {
	*rules.AbstractStatisticSentenceStyleRule
}

const passiveSentenceDefaultMinPercent = 8

func NewPassiveSentenceRule(messages map[string]string) *PassiveSentenceRule {
	r := &PassiveSentenceRule{
		AbstractStatisticSentenceStyleRule: &rules.AbstractStatisticSentenceStyleRule{
			ID:                  "PASSIVE_SENTENCE_DE",
			Description:         "Statistische Stilanalyse: Passivsätze",
			MinPercent:          0, // twin tests / Match show all; Java default 8 via UserConfig
			ExcludeDirectSpeech: true,
			Denominator:         100,
		},
	}
	r.ConditionFulfilled = r.conditionFulfilled
	r.LimitMessage = r.getLimitMessage
	rules.InitStatisticSentenceStyleMeta(r.AbstractStatisticSentenceStyleRule, messages, false)
	return r
}

// NewPassiveSentenceRuleWithDefaultLimit uses Java DEFAULT_MIN_PERCENT=8.
func NewPassiveSentenceRuleWithDefaultLimit(messages map[string]string) *PassiveSentenceRule {
	r := NewPassiveSentenceRule(messages)
	r.MinPercent = passiveSentenceDefaultMinPercent
	return r
}

func (r *PassiveSentenceRule) GetID() string {
	if r != nil && r.AbstractStatisticSentenceStyleRule != nil {
		return r.AbstractStatisticSentenceStyleRule.GetID()
	}
	return "PASSIVE_SENTENCE_DE"
}

func (r *PassiveSentenceRule) GetDescription() string {
	return "Statistische Stilanalyse: Passivsätze"
}

func (r *PassiveSentenceRule) getLimitMessage(limit int, percent float64) string {
	if limit == 0 {
		return "Passivsatz: Aktiv formulierte Sätze sprechen im Regelfall den Leser stärker an."
	}
	return "Mehr als " + itoaDE(limit) + "% Passivsätze {" + itoaDE(int(percent+0.5)) +
		"%} gefunden. Aktiv formulierte Sätze sprechen im Regelfall den Leser stärker an."
}

func (r *PassiveSentenceRule) conditionFulfilled(sentence []*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedTokenReadings {
	// Java PassiveSentenceRule.conditionFulfilled — werden lemma + VER:PA2 (either order)
	for i := 0; i < len(sentence); i++ {
		if sentence[i] == nil {
			continue
		}
		if isWerdenPassive(sentence[i]) {
			token := sentence[i]
			for j := i + 1; j < len(sentence); j++ {
				if sentence[j] == nil {
					continue
				}
				if isPassivePA2(sentence[j]) {
					return token
				} else if rules.IsStatisticMark(sentence[j]) {
					return nil
				}
			}
		} else if isPassivePA2(sentence[i]) {
			for j := i + 1; j < len(sentence); j++ {
				if sentence[j] == nil {
					continue
				}
				if isWerdenPassive(sentence[j]) {
					return sentence[j]
				} else if rules.IsStatisticMark(sentence[j]) {
					return nil
				}
			}
		}
	}
	return nil
}

func isWerdenPassive(t *languagetool.AnalyzedTokenReadings) bool {
	return t != nil && t.HasAnyLemma("werden")
}

func isPassivePA2(t *languagetool.AnalyzedTokenReadings) bool {
	return t != nil && t.HasPosTagStartingWith("VER:PA2:")
}

// Match is single-sentence convenience (Java is text-level only).
func (r *PassiveSentenceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil {
		return nil
	}
	return r.MatchList([]*languagetool.AnalyzedSentence{sentence})
}

// itoaDE minimal int string (avoid strconv import cycles / keep leaf simple).
func itoaDE(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
