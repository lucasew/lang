package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// StyleTooOftenUsedVerbRule ports org.languagetool.rules.de.StyleTooOftenUsedVerbRule.
// Java: isToCountedWord VER:; isException VER:MOD|VER:AUX|ART|ADJ (bare ART/ADJ prefixes);
// toAddedLemma VER:. DEFAULT_MIN_PERCENT=5; MIN_WORD_COUNT=100 (abstract); default off.
type StyleTooOftenUsedVerbRule struct {
	*rules.AbstractStyleTooOftenUsedWordRule
}

// styleTooOftenVerbConfigureText is Java StyleTooOftenUsedVerbRule.getConfigureText.
const styleTooOftenVerbConfigureText = "Anzeigen wenn ein Verb häufiger verwendet wird als ...% aller Verben:"

// NewStyleTooOftenUsedVerbRule ports
// StyleTooOftenUsedVerbRule(messages, lang, null).
func NewStyleTooOftenUsedVerbRule(messages map[string]string) *StyleTooOftenUsedVerbRule {
	return NewStyleTooOftenUsedVerbRuleWithUserConfig(messages, nil)
}

// NewStyleTooOftenUsedVerbRuleWithUserConfig ports
// StyleTooOftenUsedVerbRule(ResourceBundle messages, Language lang, UserConfig userConfig).
// Java: super(messages, lang, userConfig, DEFAULT_MIN_PERCENT); then if userConfig!=null and
// linguServices!=null → linguServices.setThesaurusRelevantRule(this).
// UserConfig minPercent: getConfigValueByID(getId())[0] when present (abstract getMinPercent).
func NewStyleTooOftenUsedVerbRuleWithUserConfig(messages map[string]string, userConfig *languagetool.UserConfig) *StyleTooOftenUsedVerbRule {
	minPercent := styleTooOftenDefaultMinPercent
	if userConfig != nil {
		if cf := userConfig.GetConfigValueByID("TOO_OFTEN_USED_VERB_DE"); len(cf) > 0 {
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
		ID:                "TOO_OFTEN_USED_VERB_DE",
		Description:       "Statistische Stilanalyse: Zu häufig genutztes Verb",
		ConfigureText:     styleTooOftenVerbConfigureText,
		DefaultMinPercent: styleTooOftenDefaultMinPercent,
		MinPercent:        minPercent,
		MinWordCount:      100, // Java AbstractStyleTooOftenUsedWordRule.MIN_WORD_COUNT
		// Java getLimitMessage — no space before period (unlike StyleTooOftenUsedAdjectiveRule).
		LimitMessage: func(limit int) string {
			return "Das Verb wird häufiger verwendet als " + itoaDE(limit) +
				"% aller Verben. Möglicherweise ist es besser es durch ein Synonym zu ersetzen."
		},
		// Java: isToCountedWord = hasPosTagStartingWith("VER:")
		IsToCountedWord: func(tok *languagetool.AnalyzedTokenReadings) bool {
			return tok != nil && tok.HasPosTagStartingWith("VER:")
		},
		// Java isException:
		//   hasPosTagStartingWith("VER:MOD") || "VER:AUX" || "ART" || "ADJ"
		// Critical: ART and ADJ are bare prefixes (no colon) — match ART, ART:DEF…, ADJ, ADJ:NOM…
		// exactly as Java hasPosTagStartingWith.
		IsException: func(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
			token := tokens[n]
			return token.HasPosTagStartingWith("VER:MOD") ||
				token.HasPosTagStartingWith("VER:AUX") ||
				token.HasPosTagStartingWith("ART") ||
				token.HasPosTagStartingWith("ADJ")
		},
		// Java: toAddedLemma = getLemmaForPosTagStartsWith("VER:", token)
		ToAddedLemma: func(tok *languagetool.AnalyzedTokenReadings) string {
			return rules.LemmaForPosTagStartsWith("VER:", tok)
		},
	}
	rules.InitStyleTooOftenUsedWordMeta(base, messages, false)
	r := &StyleTooOftenUsedVerbRule{AbstractStyleTooOftenUsedWordRule: base}
	// Java ctor: if (userConfig != null) { LinguServices ls = userConfig.getLinguServices();
	// if (ls != null) ls.setThesaurusRelevantRule(this); }
	if userConfig != nil {
		if ls := userConfig.GetLinguServices(); ls != nil {
			ls.SetThesaurusRelevantRule(r)
		}
	}
	return r
}

// NewStyleTooOftenUsedVerbRuleWithDefaultLimit is the register-core entry (Java defaults).
func NewStyleTooOftenUsedVerbRuleWithDefaultLimit(messages map[string]string) *StyleTooOftenUsedVerbRule {
	r := NewStyleTooOftenUsedVerbRule(messages)
	r.MinPercent = styleTooOftenDefaultMinPercent
	r.MinWordCount = 100
	return r
}

func (r *StyleTooOftenUsedVerbRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractStyleTooOftenUsedWordRule.MatchList(sentences)
}
