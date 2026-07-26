package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RussianFillerWordsRule ports org.languagetool.rules.ru.RussianFillerWordsRule
// (extends AbstractFillerWordsRule; Java default minPercent=8, default off).
type RussianFillerWordsRule struct {
	*rules.AbstractFillerWordsRule
}

// russianFillerWords is the exact static HashSet from Java RussianFillerWordsRule
// (15 tokens; case-sensitive lookup via isFillerWord).
var russianFillerWords = func() map[string]struct{} {
	words := []string{
		"ах", "аа", "ааа", "аааа", "ау",
		"бу", "вау", "ох", "однако", "эээ",
		"э", "эй", "эх", "ух-ты", "ух",
	}
	m := make(map[string]struct{}, len(words))
	for _, w := range words {
		m[w] = struct{}{}
	}
	return m
}()

// NewRussianFillerWordsRule ports RussianFillerWordsRule(messages, lang, null).
func NewRussianFillerWordsRule(messages map[string]string) *RussianFillerWordsRule {
	return NewRussianFillerWordsRuleWithUserConfig(messages, nil)
}

// NewRussianFillerWordsRuleWithUserConfig ports
// RussianFillerWordsRule(ResourceBundle messages, Language lang, UserConfig userConfig).
// Language is unused in Java beyond AbstractStatisticStyleRule storage; Go omits it.
func NewRussianFillerWordsRuleWithUserConfig(messages map[string]string, userConfig *languagetool.UserConfig) *RussianFillerWordsRule {
	base := &rules.AbstractFillerWordsRule{
		AbstractStatisticStyleRule: &rules.AbstractStatisticStyleRule{},
		Messages:                   messages,
		// Only set ID; Description/Message come from messages keys via InitFillerWordsMeta
		// (filler_words_rule_desc / filler_words_rule_msg) — do not invent RU/EN copy.
		ID:          "FILLER_WORDS_RU",
		FillerWords: russianFillerWords,
		// Java isException always returns false (German-style exception is commented out).
		IsException: russianFillerIsException,
	}
	rules.InitFillerWordsMeta(base, messages, false)

	// Java AbstractStatisticStyleRule: minPercent / excludeDirectSpeech from userConfig.
	if userConfig != nil {
		if cf := userConfig.GetConfigValueByID("FILLER_WORDS_RU"); len(cf) > 0 && cf[0] != nil {
			if n, ok := russianFillerConfigInt(cf[0]); ok {
				base.SetMinPercent(n)
			}
		}
		if cf := userConfig.GetConfigValueByID("FILLER_WORDS_RU"); len(cf) > 1 && cf[1] != nil {
			if b, ok := cf[1].(bool); ok {
				base.ExcludeDirectSpeech = b
			}
		}
	}

	r := &RussianFillerWordsRule{AbstractFillerWordsRule: base}
	// RU Java isFillerWord is case-sensitive HashSet.contains; abstract Go isFiller
	// lowercases. Override ConditionFulfilled so only exact keys match.
	base.ConditionFulfilled = r.conditionFulfilled
	return r
}

// NewRussianFillerWordsRuleWithMinPercent builds with explicit percent (0 = show all).
func NewRussianFillerWordsRuleWithMinPercent(messages map[string]string, minPercent int) *RussianFillerWordsRule {
	r := NewRussianFillerWordsRule(messages)
	r.SetMinPercent(minPercent)
	return r
}

// isFillerWord ports RussianFillerWordsRule.isFillerWord — exact, case-sensitive.
func (r *RussianFillerWordsRule) isFillerWord(token string) bool {
	if r == nil || r.FillerWords == nil {
		return false
	}
	_, ok := r.FillerWords[token]
	return ok
}

// conditionFulfilled ports AbstractFillerWordsRule.conditionFulfilled with RU isFillerWord.
func (r *RussianFillerWordsRule) conditionFulfilled(tokens []*languagetool.AnalyzedTokenReadings, n int) int {
	if r == nil || n < 0 || n >= len(tokens) || tokens[n] == nil {
		return -1
	}
	if r.isFillerWord(tokens[n].GetToken()) && !russianFillerIsException(tokens, n) {
		return n
	}
	return -1
}

// russianFillerIsException ports RussianFillerWordsRule.isException — always false.
func russianFillerIsException(tokens []*languagetool.AnalyzedTokenReadings, num int) bool {
	// Java has a commented German-style "aber"+comma exception; active body is return false.
	return false
}

func russianFillerConfigInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}

func (r *RussianFillerWordsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractFillerWordsRule == nil {
		return nil
	}
	return r.AbstractFillerWordsRule.Match(sentence)
}

func (r *RussianFillerWordsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractFillerWordsRule == nil {
		return nil
	}
	return r.AbstractFillerWordsRule.MatchList(sentences)
}
