package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// StyleTooOftenUsedAdjectiveRule ports org.languagetool.rules.de.StyleTooOftenUsedAdjectiveRule.
// Java: isToCountedWord ADJ:; isException PRO:|ADV:|ZUS; toAddedLemma ADJ:.
// Java DEFAULT_MIN_PERCENT=5; MIN_WORD_COUNT=100 (abstract); default off.
type StyleTooOftenUsedAdjectiveRule struct {
	*rules.AbstractStyleTooOftenUsedWordRule
}

// styleTooOftenAdjConfigureText is Java StyleTooOftenUsedAdjectiveRule.getConfigureText.
const styleTooOftenAdjConfigureText = "Anzeigen wenn ein Adjektiv häufiger verwendet wird als ...% aller Adjektive:"

// NewStyleTooOftenUsedAdjectiveRule ports
// StyleTooOftenUsedAdjectiveRule(messages, lang, null).
func NewStyleTooOftenUsedAdjectiveRule(messages map[string]string) *StyleTooOftenUsedAdjectiveRule {
	return NewStyleTooOftenUsedAdjectiveRuleWithUserConfig(messages, nil)
}

// NewStyleTooOftenUsedAdjectiveRuleWithUserConfig ports
// StyleTooOftenUsedAdjectiveRule(ResourceBundle messages, Language lang, UserConfig userConfig).
// Java: super(messages, lang, userConfig, DEFAULT_MIN_PERCENT); then if userConfig!=null and
// linguServices!=null → linguServices.setThesaurusRelevantRule(this).
// UserConfig minPercent: getConfigValueByID(getId())[0] when present (abstract getMinPercent).
func NewStyleTooOftenUsedAdjectiveRuleWithUserConfig(messages map[string]string, userConfig *languagetool.UserConfig) *StyleTooOftenUsedAdjectiveRule {
	minPercent := styleTooOftenDefaultMinPercent
	if userConfig != nil {
		if cf := userConfig.GetConfigValueByID("TOO_OFTEN_USED_ADJECTIVE_DE"); len(cf) > 0 {
			switch v := cf[0].(type) {
			case int:
				minPercent = v
			case int32:
				minPercent = int(v)
			case int64:
				minPercent = int(v)
			case float64:
				// Java (int) on boxed Number may arrive as float in JSON-ish configs.
				minPercent = int(v)
			}
		}
	}
	base := &rules.AbstractStyleTooOftenUsedWordRule{
		Messages:          messages,
		ID:                "TOO_OFTEN_USED_ADJECTIVE_DE",
		Description:       "Statistische Stilanalyse: Zu häufig genutztes Adjektiv",
		ConfigureText:     styleTooOftenAdjConfigureText,
		DefaultMinPercent: styleTooOftenDefaultMinPercent,
		MinPercent:        minPercent,
		MinWordCount:      100, // Java AbstractStyleTooOftenUsedWordRule.MIN_WORD_COUNT
		// Java getLimitMessage — note space before period: "Adjektive . "
		LimitMessage: func(limit int) string {
			return "Das Adjektiv wird häufiger verwendet als " + itoaDE(limit) +
				"% aller Adjektive . Möglicherweise ist es besser es durch ein Synonym zu ersetzen."
		},
		// Java: isToCountedWord = hasPosTagStartingWith("ADJ:")
		IsToCountedWord: func(tok *languagetool.AnalyzedTokenReadings) bool {
			return tok != nil && tok.HasPosTagStartingWith("ADJ:")
		},
		// Java: isException = PRO: || ADV: || ZUS
		IsException: func(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
			token := tokens[n]
			return token.HasPosTagStartingWith("PRO:") ||
				token.HasPosTagStartingWith("ADV:") ||
				token.HasPosTagStartingWith("ZUS")
		},
		// Java: toAddedLemma = getLemmaForPosTagStartsWith("ADJ:", token)
		ToAddedLemma: func(tok *languagetool.AnalyzedTokenReadings) string {
			return rules.LemmaForPosTagStartsWith("ADJ:", tok)
		},
	}
	rules.InitStyleTooOftenUsedWordMeta(base, messages, false)
	r := &StyleTooOftenUsedAdjectiveRule{AbstractStyleTooOftenUsedWordRule: base}
	// Java ctor: if (userConfig != null) { LinguServices ls = userConfig.getLinguServices();
	// if (ls != null) ls.setThesaurusRelevantRule(this); }
	if userConfig != nil {
		if ls := userConfig.GetLinguServices(); ls != nil {
			ls.SetThesaurusRelevantRule(r)
		}
	}
	return r
}

// NewStyleTooOftenUsedAdjectiveRuleWithDefaultLimit is the register-core entry (Java defaults).
func NewStyleTooOftenUsedAdjectiveRuleWithDefaultLimit(messages map[string]string) *StyleTooOftenUsedAdjectiveRule {
	r := NewStyleTooOftenUsedAdjectiveRule(messages)
	r.MinPercent = styleTooOftenDefaultMinPercent
	r.MinWordCount = 100
	return r
}

func (r *StyleTooOftenUsedAdjectiveRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractStyleTooOftenUsedWordRule.MatchList(sentences)
}
