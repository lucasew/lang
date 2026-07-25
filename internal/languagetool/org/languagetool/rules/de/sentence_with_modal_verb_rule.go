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

// Java DEFAULT_MIN_PERCENT
const sentenceWithModalDefaultMinPercent = 18

// Java getConfigureText
const sentenceWithModalConfigureText = "Anzeigen wenn mehr als ...% Sätze eines Kapitels Modalverben enthalten:"

// NewSentenceWithModalVerbRule ports
// SentenceWithModalVerbRule(ResourceBundle messages, Language lang, UserConfig userConfig)
// with null UserConfig (DEFAULT_MIN_PERCENT=18).
func NewSentenceWithModalVerbRule(messages map[string]string) *SentenceWithModalVerbRule {
	return NewSentenceWithModalVerbRuleWithUserConfig(messages, nil)
}

// NewSentenceWithModalVerbRuleWithUserConfig ports
// SentenceWithModalVerbRule(messages, lang, userConfig) → super(..., DEFAULT_MIN_PERCENT)
// with minPercent from UserConfig.getConfigValueByID(getId()) when present.
func NewSentenceWithModalVerbRuleWithUserConfig(messages map[string]string, userConfig *languagetool.UserConfig) *SentenceWithModalVerbRule {
	minPercent := sentenceWithModalDefaultMinPercent
	if userConfig != nil {
		if cf := userConfig.GetConfigValueByID("SENTENCE_WITH_MODAL_VERB_DE"); len(cf) > 0 && cf[0] != nil {
			if n, ok := configInt(cf[0]); ok {
				minPercent = n
			}
		}
	}
	r := &SentenceWithModalVerbRule{
		AbstractStatisticSentenceStyleRule: &rules.AbstractStatisticSentenceStyleRule{
			ID:                  "SENTENCE_WITH_MODAL_VERB_DE",
			Description:         "Statistische Stilanalyse: Sätze mit Modalverben",
			MinPercent:          minPercent,
			ExcludeDirectSpeech: true, // Java excludeDirectSpeech()
			Denominator:         100,  // Java denominator() default
		},
	}
	r.ConditionFulfilled = r.conditionFulfilled
	r.LimitMessage = r.getLimitMessage
	rules.InitStatisticSentenceStyleMeta(r.AbstractStatisticSentenceStyleRule, messages, false)
	return r
}

// NewSentenceWithModalVerbRuleWithDefaultLimit is an alias of the Java constructor (18%).
func NewSentenceWithModalVerbRuleWithDefaultLimit(messages map[string]string) *SentenceWithModalVerbRule {
	return NewSentenceWithModalVerbRule(messages)
}

// NewSentenceWithModalVerbRuleWithMinPercent sets an explicit limit (tests / UserConfig override).
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

// GetConfigureText ports SentenceWithModalVerbRule.getConfigureText.
func (r *SentenceWithModalVerbRule) GetConfigureText() string {
	return sentenceWithModalConfigureText
}

// GetRuleOptions ports AbstractStatisticSentenceStyleRule.getRuleOptions via subclass configure text.
// Java: new RuleOption(defaultMinPercent, getConfigureText(), 0, 100)
func (r *SentenceWithModalVerbRule) GetRuleOptions() []*rules.RuleOption {
	return []*rules.RuleOption{
		rules.NewRuleOption(sentenceWithModalDefaultMinPercent, r.GetConfigureText(), 0, 100),
	}
}

func (r *SentenceWithModalVerbRule) getLimitMessage(limit int, percent float64) string {
	// Java SentenceWithModalVerbRule.getLimitMessage
	if limit == 0 {
		return "Modalverb: Modalverben blähen den Text häufig auf und sollten vermieden werden."
	}
	return "Mehr als " + itoaDE(limit) + "% Sätze mit Modalverben {" + itoaDE(int(percent+0.5)) +
		"%} gefunden. Modalverben blähen den Text häufig auf und sollten vermieden werden."
}

// isModalVerb ports private static isModalVerb — hasPosTagStartingWith("VER:MOD") only.
func (r *SentenceWithModalVerbRule) isModalVerb(token *languagetool.AnalyzedTokenReadings) bool {
	return token != nil && token.HasPosTagStartingWith("VER:MOD")
}

func (r *SentenceWithModalVerbRule) conditionFulfilled(sentence []*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedTokenReadings {
	// Java SentenceWithModalVerbRule.conditionFulfilled — reuses loop index i in inner for (i++; …).
	// MOD then later INF → return MOD token; INF then later MOD → return MOD token; isMark → null.
	for i := 0; i < len(sentence); i++ {
		if r.isModalVerb(sentence[i]) {
			token := sentence[i]
			for i++; i < len(sentence); i++ {
				if sentence[i] != nil && sentence[i].HasPosTagStartingWith("VER:INF") {
					return token
				} else if rules.IsStatisticMark(sentence[i]) {
					return nil
				}
			}
		} else if sentence[i] != nil && sentence[i].HasPosTagStartingWith("VER:INF") {
			for i++; i < len(sentence); i++ {
				if r.isModalVerb(sentence[i]) {
					return sentence[i]
				} else if rules.IsStatisticMark(sentence[i]) {
					return nil
				}
			}
		}
	}
	return nil
}

// Match single-sentence convenience (Java is text-level MatchList).
func (r *SentenceWithModalVerbRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil {
		return nil
	}
	return r.MatchList([]*languagetool.AnalyzedSentence{sentence})
}
