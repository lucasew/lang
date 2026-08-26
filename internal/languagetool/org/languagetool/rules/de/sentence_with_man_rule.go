package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SentenceWithManRule ports org.languagetool.rules.de.SentenceWithManRule
// (AbstractStatisticSentenceStyleRule; default off; DEFAULT_MIN_PERCENT=15; ‰ via denominator 1000).
// Java: hasLemma("man") only — no surface form invent.
type SentenceWithManRule struct {
	*rules.AbstractStatisticSentenceStyleRule
}

// Java DEFAULT_MIN_PERCENT
const sentenceWithManDefaultMinPercent = 15

// Java getConfigureText (unit is per mil / ‰)
const sentenceWithManConfigureText = "Anzeigen wenn mehr als ...‰ Sätze eines Kapitels 'man' enthalten:"

// NewSentenceWithManRule ports
// SentenceWithManRule(ResourceBundle messages, Language lang, UserConfig userConfig)
// with null UserConfig (DEFAULT_MIN_PERCENT=15).
func NewSentenceWithManRule(messages map[string]string) *SentenceWithManRule {
	return NewSentenceWithManRuleWithUserConfig(messages, nil)
}

// NewSentenceWithManRuleWithUserConfig ports
// SentenceWithManRule(messages, lang, userConfig) → super(..., DEFAULT_MIN_PERCENT)
// with minPercent from UserConfig.getConfigValueByID(getId()) when present.
func NewSentenceWithManRuleWithUserConfig(messages map[string]string, userConfig *languagetool.UserConfig) *SentenceWithManRule {
	minPercent := sentenceWithManDefaultMinPercent
	if userConfig != nil {
		if cf := userConfig.GetConfigValueByID("SENTENCE_WITH_MAN_DE"); len(cf) > 0 && cf[0] != nil {
			if n, ok := configInt(cf[0]); ok {
				minPercent = n
			}
		}
	}
	r := &SentenceWithManRule{
		AbstractStatisticSentenceStyleRule: &rules.AbstractStatisticSentenceStyleRule{
			ID:                  "SENTENCE_WITH_MAN_DE",
			Description:         "Statistische Stilanalyse: Sätze mit indirekter Leseransprache 'man'",
			MinPercent:          minPercent,
			ExcludeDirectSpeech: true, // Java excludeDirectSpeech()
			Denominator:         1000, // Java denominator() → 1000.0 (messages use ‰)
		},
	}
	r.ConditionFulfilled = r.conditionFulfilled
	r.LimitMessage = r.getLimitMessage
	rules.InitStatisticSentenceStyleMeta(r.AbstractStatisticSentenceStyleRule, messages, false)
	return r
}

// NewSentenceWithManRuleWithDefaultLimit is an alias of the Java constructor (15‰).
func NewSentenceWithManRuleWithDefaultLimit(messages map[string]string) *SentenceWithManRule {
	return NewSentenceWithManRule(messages)
}

// NewSentenceWithManRuleWithMinPercent sets an explicit limit (tests / UserConfig override).
func NewSentenceWithManRuleWithMinPercent(messages map[string]string, min int) *SentenceWithManRule {
	r := NewSentenceWithManRule(messages)
	r.MinPercent = min
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

// GetConfigureText ports SentenceWithManRule.getConfigureText.
func (r *SentenceWithManRule) GetConfigureText() string {
	return sentenceWithManConfigureText
}

// GetRuleOptions ports AbstractStatisticSentenceStyleRule.getRuleOptions via subclass configure text.
// Java: new RuleOption(defaultMinPercent, getConfigureText(), 0, 100)
func (r *SentenceWithManRule) GetRuleOptions() []*rules.RuleOption {
	return []*rules.RuleOption{
		rules.NewRuleOption(sentenceWithManDefaultMinPercent, r.GetConfigureText(), 0, 100),
	}
}

func (r *SentenceWithManRule) getLimitMessage(limit int, percent float64) string {
	// Java SentenceWithManRule.getLimitMessage
	if limit == 0 {
		return "Sätze mit der indirekten Leseransprache 'man' sind stilistisch wenig elegant formuliert. Lässt sich das Wort vermeiden?"
	}
	return "Mehr als " + itoaDE(limit) + "‰ Sätze mit der indirekten Leseransprache 'man' {" +
		itoaDE(int(percent+0.5)) + "‰} gefunden. Lässt sich das Wort vermeiden?"
}

// isWordMan ports private static isWordMan — hasLemma("man") only.
func isWordMan(token *languagetool.AnalyzedTokenReadings) bool {
	return token != nil && token.HasLemma("man")
}

func (r *SentenceWithManRule) conditionFulfilled(sentence []*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedTokenReadings {
	// Java SentenceWithManRule.conditionFulfilled — first token with lemma "man".
	for i := 0; i < len(sentence); i++ {
		if isWordMan(sentence[i]) {
			return sentence[i]
		}
	}
	return nil
}

// Match single-sentence convenience (Java is text-level MatchList).
func (r *SentenceWithManRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil {
		return nil
	}
	return r.MatchList([]*languagetool.AnalyzedSentence{sentence})
}
