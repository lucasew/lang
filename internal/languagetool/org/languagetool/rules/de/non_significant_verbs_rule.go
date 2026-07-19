package de

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// NonSignificantVerbsRule ports org.languagetool.rules.de.NonSignificantVerbsRule
// (AbstractStatisticStyleRule; per-mill; default off; DEFAULT_MIN_PER_MILL=8).
// Java: hasAnyLemma(haben|sein|machen|tun) only — no surface form invent.
type NonSignificantVerbsRule struct {
	*rules.AbstractStatisticStyleRule
}

var nonSignificantLemmas = []string{"haben", "sein", "machen", "tun"}

const nonSignificantDefaultMinPerMill = 8

func NewNonSignificantVerbsRule(messages map[string]string) *NonSignificantVerbsRule {
	r := &NonSignificantVerbsRule{
		AbstractStatisticStyleRule: &rules.AbstractStatisticStyleRule{
			ID:                  "NON_SIGNIFICANT_VERB_DE",
			Description:         "Statistische Stilanalyse: Verben mit wenig Aussagekraft",
			MinPercent:          0, // twin tests show all; Java default 8‰
			Denominator:         1000,
			ExcludeDirectSpeech: true,
		},
	}
	r.ConditionFulfilled = r.conditionFulfilled
	r.SentenceConditionFulfilled = func(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
		return false
	}
	r.LimitMessage = r.getLimitMessage
	rules.InitStatisticStyleMeta(r.AbstractStatisticStyleRule, messages, false)
	// Java: machte … / fixed is a different sentence (illustrative)
	r.AddExamplePair(
		rules.Wrong("Er <marker>machte</marker> einen Kuchen."),
		rules.Fixed("Das macht mir Angst."),
	)
	return r
}

func NewNonSignificantVerbsRuleWithDefaultLimit(messages map[string]string) *NonSignificantVerbsRule {
	r := NewNonSignificantVerbsRule(messages)
	r.MinPercent = nonSignificantDefaultMinPerMill
	return r
}

func (r *NonSignificantVerbsRule) GetID() string {
	if r != nil && r.AbstractStatisticStyleRule != nil {
		return r.AbstractStatisticStyleRule.GetID()
	}
	return "NON_SIGNIFICANT_VERB_DE"
}

func (r *NonSignificantVerbsRule) GetDescription() string {
	return "Statistische Stilanalyse: Verben mit wenig Aussagekraft"
}

func (r *NonSignificantVerbsRule) getLimitMessage(limit int, percent float64) string {
	if limit == 0 {
		return "Dieses Verb hat wenig Aussagekraft. Verwenden Sie wenn möglich ein anderes oder formulieren Sie den Satz um."
	}
	return "Mehr als " + itoaDE(limit) + "‰ wenig aussagekräftige Verben {" + itoaDE(int(percent+0.5)) +
		"‰} gefunden. Verwenden Sie wenn möglich ein anderes Verb oder formulieren Sie den Satz um."
}

func (r *NonSignificantVerbsRule) conditionFulfilled(tokens []*languagetool.AnalyzedTokenReadings, n int) int {
	if n < 0 || n >= len(tokens) || tokens[n] == nil {
		return -1
	}
	// Java: hasAnyLemma(nonSignificant) && !isException
	if tokens[n].HasAnyLemma(nonSignificantLemmas...) && !isNonSignificantException(tokens, n) {
		return n
	}
	return -1
}

func isNonSignificantException(tokens []*languagetool.AnalyzedTokenReadings, num int) bool {
	if tokens[num] == nil {
		return true
	}
	surface := tokens[num].GetToken()
	if strings.HasPrefix(surface, "sein") || strings.HasPrefix(surface, "Sein") {
		return true
	}
	if tokens[num].HasAnyLemma("machen") {
		for i := 1; i < len(tokens); i++ {
			if tokens[i] == nil {
				continue
			}
			s := tokens[i].GetToken()
			switch s {
			case "Angst", "Weg", "frisch", "bemerkbar", "aufmerksam":
				return true
			}
		}
		return false
	}
	isHaben := tokens[num].HasAnyLemma("haben")
	if isHaben {
		for i := 1; i < len(tokens); i++ {
			if tokens[i] == nil {
				continue
			}
			s := tokens[i].GetToken()
			if s == "Glück" || s == "Angst" || s == "Mühe" || s == "Recht" || s == "recht" {
				return true
			}
		}
	}
	if isHaben || tokens[num].HasAnyLemma("sein") {
		for i := 1; i < len(tokens); i++ {
			if tokens[i] == nil {
				continue
			}
			if tokens[i].HasPosTagStartingWith("PA2") || tokens[i].HasPosTagStartingWith("VER:PA2") ||
				tokens[i].GetToken() == "Flucht" || isUnknownWordNS(tokens[i]) {
				return true
			}
		}
	}
	return false
}

// isUnknownWordNS ports NonSignificantVerbsRule.isUnknownWord.
// Java: isPosTagUnknown() && len>2 && letters only.
// isPosTagUnknown = single reading with null POS (Go: !IsTagged()).
func isUnknownWordNS(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil || token.IsTagged() {
		return false
	}
	s := token.GetToken()
	if len([]rune(s)) <= 2 {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// Match single-sentence convenience (Java text-level MatchList).
func (r *NonSignificantVerbsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil {
		return nil
	}
	return r.MatchList([]*languagetool.AnalyzedSentence{sentence})
}
