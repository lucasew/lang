package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ConjunctionAtBeginOfSentenceRule ports
// org.languagetool.rules.de.ConjunctionAtBeginOfSentenceRule
// (AbstractStatisticSentenceStyleRule; default off; DEFAULT_MIN_PERCENT=10).
// Java: hasPosTagStartingWith("KON") only — fillerWords list is commented out (no invent).
type ConjunctionAtBeginOfSentenceRule struct {
	*rules.AbstractStatisticSentenceStyleRule
}

const conjunctionBeginDefaultMinPercent = 10

func NewConjunctionAtBeginOfSentenceRule(messages map[string]string) *ConjunctionAtBeginOfSentenceRule {
	r := &ConjunctionAtBeginOfSentenceRule{
		AbstractStatisticSentenceStyleRule: &rules.AbstractStatisticSentenceStyleRule{
			// Java getId
			ID:                  "SENTENCE_BEGINNING_WITH_CONJUNCTION_DE",
			Description:         "Statistische Stilanalyse: Sätze beginnend mit Konjunktion",
			MinPercent:          10, // Java DEFAULT_MIN_PERCENT
			ExcludeDirectSpeech: true,
			Denominator:         100,
		},
	}
	r.ConditionFulfilled = r.conditionFulfilled
	r.LimitMessage = r.getLimitMessage
	rules.InitStatisticSentenceStyleMeta(r.AbstractStatisticSentenceStyleRule, messages, false)
	return r
}

func NewConjunctionAtBeginOfSentenceRuleWithDefaultLimit(messages map[string]string) *ConjunctionAtBeginOfSentenceRule {
	r := NewConjunctionAtBeginOfSentenceRule(messages)
	r.MinPercent = conjunctionBeginDefaultMinPercent
	return r
}

func (r *ConjunctionAtBeginOfSentenceRule) GetID() string {
	if r != nil && r.AbstractStatisticSentenceStyleRule != nil {
		return r.AbstractStatisticSentenceStyleRule.GetID()
	}
	return "SENTENCE_BEGINNING_WITH_CONJUNCTION_DE"
}

func (r *ConjunctionAtBeginOfSentenceRule) GetDescription() string {
	return "Statistische Stilanalyse: Sätze beginnend mit Konjunktion"
}

func (r *ConjunctionAtBeginOfSentenceRule) getLimitMessage(limit int, percent float64) string {
	if limit == 0 {
		return "Eine Konjunktion sollte nur in Ausnahmefällen am Satzanfang verwendet werden. Formulieren Sie den Satz um, falls möglich."
	}
	return "Mehr als " + itoaDE(limit) + "% Sätze beginnen mit einer Konjunktion {" +
		itoaDE(int(percent+0.5)) + "%} gefunden. Formulieren Sie den Satz um, falls möglich."
}

func isConjunctionToken(token *languagetool.AnalyzedTokenReadings) bool {
	// Java: hasPosTagStartingWith("KON") only
	return token != nil && token.HasPosTagStartingWith("KON")
}

func isCommaToken(token *languagetool.AnalyzedTokenReadings) bool {
	return token != nil && token.GetToken() == ","
}

func (r *ConjunctionAtBeginOfSentenceRule) conditionFulfilled(sentence []*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedTokenReadings {
	// Java works on relevant sentence part (no SENT_START); list index 0 is first word.
	if len(sentence) < 3 {
		return nil
	}
	num := 0
	if rules.IsStatisticOpeningQuote(sentence[0]) {
		num++
	}
	if num >= len(sentence) {
		return nil
	}
	var token *languagetool.AnalyzedTokenReadings
	if isConjunctionToken(sentence[num]) {
		token = sentence[num]
	}
	if token == nil {
		return nil
	}
	// Java exceptions by surface
	if token.GetToken() == "Wie" || token.GetToken() == "Seit" || token.GetToken() == "Allerdings" ||
		(token.GetToken() == "Aber" && num+1 < len(sentence) && sentence[num+1] != nil && sentence[num+1].GetToken() == "auch") {
		return nil
	}
	if token.GetToken() == "Um" {
		for i := 1; i < len(sentence); i++ {
			if isCommaToken(sentence[i]) || (sentence[i] != nil && sentence[i].GetToken() == "herum") {
				return nil
			}
		}
		return token
	}
	// KON:UNT path vs other (Java)
	if !token.HasPosTagStartingWith("KON:UNT") || token.GetToken() == "Sondern" ||
		(token.GetToken() == "Auch" && num+1 < len(sentence) && sentence[num+1] != nil && sentence[num+1].GetToken() == "wenn") {
		if token.GetToken() == "Entweder" {
			for i := 1; i < len(sentence); i++ {
				if sentence[i] != nil && sentence[i].GetToken() == "oder" {
					return nil
				}
			}
		} else if token.GetToken() == "Sowohl" {
			for i := 1; i < len(sentence)-1; i++ {
				if sentence[i] != nil && sentence[i+1] != nil &&
					sentence[i].GetToken() == "als" && sentence[i+1].GetToken() == "auch" {
					return nil
				}
			}
		} else if token.GetToken() == "Weder" {
			for i := 1; i < len(sentence); i++ {
				if sentence[i] != nil && sentence[i].GetToken() == "noch" {
					return nil
				}
			}
		} else {
			// not Entweder/Sowohl/Weder: if ? at end, null; else return token
			if sentence[len(sentence)-1] != nil && sentence[len(sentence)-1].GetToken() == "?" {
				return nil
			}
			return token
		}
	}
	// KON:UNT and not Sondern/Auch wenn: scan for comma after pos 2
	for i := 2; i < len(sentence); i++ {
		if isCommaToken(sentence[i]) {
			return nil
		}
	}
	return token
}

func (r *ConjunctionAtBeginOfSentenceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil {
		return nil
	}
	return r.MatchList([]*languagetool.AnalyzedSentence{sentence})
}
