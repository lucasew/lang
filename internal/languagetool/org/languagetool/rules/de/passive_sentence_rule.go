package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PassiveSentenceRule ports org.languagetool.rules.de.PassiveSentenceRule
// (AbstractStatisticSentenceStyleRule; default off; DEFAULT_MIN_PERCENT=8).
// Java: hasLemma("werden") + hasPosTagStartingWith("VER:PA2:") either order — no surface invent.
type PassiveSentenceRule struct {
	*rules.AbstractStatisticSentenceStyleRule
}

// Java DEFAULT_MIN_PERCENT
const passiveSentenceDefaultMinPercent = 8

// Java getConfigureText
const passiveSentenceConfigureText = "Anzeigen wenn mehr als ...% Sätze eines Kapitels Passivsätze sind:"

// NewPassiveSentenceRule ports
// PassiveSentenceRule(ResourceBundle messages, Language lang, UserConfig userConfig)
// with null UserConfig (DEFAULT_MIN_PERCENT=8).
func NewPassiveSentenceRule(messages map[string]string) *PassiveSentenceRule {
	return NewPassiveSentenceRuleWithUserConfig(messages, nil)
}

// NewPassiveSentenceRuleWithUserConfig ports
// PassiveSentenceRule(messages, lang, userConfig) → super(..., DEFAULT_MIN_PERCENT)
// with minPercent from UserConfig.getConfigValueByID(getId()) when present.
func NewPassiveSentenceRuleWithUserConfig(messages map[string]string, userConfig *languagetool.UserConfig) *PassiveSentenceRule {
	minPercent := passiveSentenceDefaultMinPercent
	if userConfig != nil {
		if cf := userConfig.GetConfigValueByID("PASSIVE_SENTENCE_DE"); len(cf) > 0 && cf[0] != nil {
			if n, ok := configInt(cf[0]); ok {
				minPercent = n
			}
		}
	}
	r := &PassiveSentenceRule{
		AbstractStatisticSentenceStyleRule: &rules.AbstractStatisticSentenceStyleRule{
			ID:                  "PASSIVE_SENTENCE_DE",
			Description:         "Statistische Stilanalyse: Passivsätze",
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

// NewPassiveSentenceRuleWithDefaultLimit is an alias of the Java constructor (8%).
func NewPassiveSentenceRuleWithDefaultLimit(messages map[string]string) *PassiveSentenceRule {
	return NewPassiveSentenceRule(messages)
}

// NewPassiveSentenceRuleWithMinPercent sets an explicit limit (tests / UserConfig override).
func NewPassiveSentenceRuleWithMinPercent(messages map[string]string, min int) *PassiveSentenceRule {
	r := NewPassiveSentenceRule(messages)
	r.MinPercent = min
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

// GetConfigureText ports PassiveSentenceRule.getConfigureText.
func (r *PassiveSentenceRule) GetConfigureText() string {
	return passiveSentenceConfigureText
}

// GetRuleOptions ports AbstractStatisticSentenceStyleRule.getRuleOptions via subclass configure text.
// Java: new RuleOption(defaultMinPercent, getConfigureText(), 0, 100)
func (r *PassiveSentenceRule) GetRuleOptions() []*rules.RuleOption {
	return []*rules.RuleOption{
		rules.NewRuleOption(passiveSentenceDefaultMinPercent, r.GetConfigureText(), 0, 100),
	}
}

func (r *PassiveSentenceRule) getLimitMessage(limit int, percent float64) string {
	// Java PassiveSentenceRule.getLimitMessage
	if limit == 0 {
		return "Passivsatz: Aktiv formulierte Sätze sprechen im Regelfall den Leser stärker an."
	}
	return "Mehr als " + itoaDE(limit) + "% Passivsätze {" + itoaDE(int(percent+0.5)) +
		"%} gefunden. Aktiv formulierte Sätze sprechen im Regelfall den Leser stärker an."
}

// hasWerdenLemma ports Java hasLemma("werden") — lemma only, no surface invent.
func hasWerdenLemma(token *languagetool.AnalyzedTokenReadings) bool {
	return token != nil && token.HasLemma("werden")
}

// hasPassivePA2 ports Java hasPosTagStartingWith("VER:PA2:").
func hasPassivePA2(token *languagetool.AnalyzedTokenReadings) bool {
	return token != nil && token.HasPosTagStartingWith("VER:PA2:")
}

func (r *PassiveSentenceRule) conditionFulfilled(sentence []*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedTokenReadings {
	// Java PassiveSentenceRule.conditionFulfilled — reuses loop index i in inner for (i++; …).
	// werden then later VER:PA2: → return werden token;
	// VER:PA2: then later werden → return werden token;
	// isMark → null.
	for i := 0; i < len(sentence); i++ {
		if hasWerdenLemma(sentence[i]) {
			token := sentence[i]
			for i++; i < len(sentence); i++ {
				if hasPassivePA2(sentence[i]) {
					return token
				} else if rules.IsStatisticMark(sentence[i]) {
					return nil
				}
			}
		} else if hasPassivePA2(sentence[i]) {
			for i++; i < len(sentence); i++ {
				if hasWerdenLemma(sentence[i]) {
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
func (r *PassiveSentenceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil {
		return nil
	}
	return r.MatchList([]*languagetool.AnalyzedSentence{sentence})
}

// itoaDE minimal int string (package-local helper used by statistic de rules).
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
