package de

import (
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// UnnecessaryPhraseRule ports org.languagetool.rules.de.UnnecessaryPhraseRule
// (AbstractStatisticStyleRule; per 10_000; default off).
type UnnecessaryPhraseRule struct {
	*rules.AbstractStatisticStyleRule
	phrases [][]string
}

const unnecessaryPhraseDefaultMinPerMill = 8

func NewUnnecessaryPhraseRule(messages map[string]string) *UnnecessaryPhraseRule {
	phrases := [][]string{
		{"dann", "und", "wann"},
		{"des", "Ungeachtet"},
		{"ganz", "und", "gar"},
		{"hie", "und", "da"},
		{"im", "Allgemeinen"},
		{"in", "der", "Tat"},
		{"in", "diesem", "Zusammenhang"},
		{"mehr", "oder", "weniger"},
		{"meines", "Erachtens"},
		{"ohne", "weiteres"},
		{"ohne", "Zweifel"},
		{"samt", "und", "sonders"},
		{"sowohl", "als", "auch"},
		{"voll", "und", "ganz"},
		{"von", "Neuem"},
		{"allem", "Anschein", "nach"},
		{"aufs", "Neue"},
		{"ein", "bisschen"},
		{"ein", "wenig"},
		{"des", "Öfteren"},
		{"bei", "weitem"},
		{"an", "sich"},
	}
	r := &UnnecessaryPhraseRule{
		AbstractStatisticStyleRule: &rules.AbstractStatisticStyleRule{
			// Java getId: UNNECESSARY_PHRASES_DE
			ID:                  "UNNECESSARY_PHRASES_DE",
			Description:         "Statistische Stilanalyse: Potenzielle Phrasen",
			MinPercent:          8, // Java DEFAULT_MIN_PER_MILL
			Denominator:         10000,
			ExcludeDirectSpeech: true,
		},
		phrases: phrases,
	}
	r.ConditionFulfilled = r.conditionFulfilled
	r.SentenceConditionFulfilled = func(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
		return false
	}
	r.LimitMessage = r.getLimitMessage
	rules.InitStatisticStyleMeta(r.AbstractStatisticStyleRule, messages, false)
	// Java: allem Anschein nach (fixed has no marker)
	r.AddExamplePair(
		rules.Wrong("Das ist <marker>allem Anschein nach</marker> eine Phrase."),
		rules.Fixed("Das ist eine Phrase."),
	)
	return r
}

func NewUnnecessaryPhraseRuleWithDefaultLimit(messages map[string]string) *UnnecessaryPhraseRule {
	r := NewUnnecessaryPhraseRule(messages)
	r.MinPercent = unnecessaryPhraseDefaultMinPerMill
	return r
}

func (r *UnnecessaryPhraseRule) GetID() string {
	if r != nil && r.AbstractStatisticStyleRule != nil {
		return r.AbstractStatisticStyleRule.GetID()
	}
	return "UNNECESSARY_PHRASES_DE"
}

func (r *UnnecessaryPhraseRule) GetDescription() string {
	return "Statistische Stilanalyse: Potenzielle Phrasen"
}

func (r *UnnecessaryPhraseRule) getLimitMessage(limit int, percent float64) string {
	if limit == 0 {
		return "Der Ausdruck gilt als Phrase. Es wird empfohlen ihn zu löschen, falls möglich."
	}
	return "Mehr als " + itoaDE(limit) + "‱ potenzielle Phrasen {" + itoaDE(int(percent+0.5)) +
		"‱} gefunden. Es wird empfohlen den Ausdruck zu löschen, falls möglich."
}

func firstCharToLowerPhrase(tokens []*languagetool.AnalyzedTokenReadings, nToken int) string {
	if nToken < 0 || nToken >= len(tokens) || tokens[nToken] == nil {
		return ""
	}
	token := tokens[nToken].GetToken()
	// Java UnnecessaryPhraseRule.firstCharToLower:
	// (nToken != 1 || token.length() < 2) ? token
	//   : token.substring(0, 1).toLowerCase() + token.substring(1)
	// length/substring are UTF-16 units — not Go UTF-8 bytes (breaks Ä/Ö/Ü).
	if nToken != 1 || utf16LenDE(token) < 2 {
		return token
	}
	u := utf16.Encode([]rune(token))
	first := string(utf16.Decode(u[:1]))
	rest := string(utf16.Decode(u[1:]))
	return strings.ToLower(first) + rest
}

func isUnnecessaryPhraseException(tokens []*languagetool.AnalyzedTokenReadings, phrase []string) bool {
	if len(phrase) == 2 && phrase[0] == "an" && phrase[1] == "sich" {
		for j := 1; j < len(tokens); j++ {
			if tokens[j] != nil && (tokens[j].HasAnyLemma("drücken") ||
				strings.HasPrefix(strings.ToLower(tokens[j].GetToken()), "drück")) {
				return true
			}
		}
	}
	return false
}

func (r *UnnecessaryPhraseRule) conditionFulfilled(tokens []*languagetool.AnalyzedTokenReadings, n int) int {
	for _, phrase := range r.phrases {
		j := 0
		for j < len(phrase) && n+j < len(tokens) &&
			phrase[j] == firstCharToLowerPhrase(tokens, n+j) {
			j++
		}
		if j == len(phrase) {
			if isUnnecessaryPhraseException(tokens, phrase) {
				return -1
			}
			return n + len(phrase) - 1
		}
	}
	return -1
}

// Match single-sentence convenience (Java is text-level).
func (r *UnnecessaryPhraseRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil {
		return nil
	}
	return r.MatchList([]*languagetool.AnalyzedSentence{sentence})
}

func (r *UnnecessaryPhraseRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractStatisticStyleRule == nil {
		return nil
	}
	return r.AbstractStatisticStyleRule.MatchList(sentences)
}
