package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// StyleTooOftenUsedNounRule ports org.languagetool.rules.de.StyleTooOftenUsedNounRule.
// Java: isToCountedWord SUB:; isException PRO:|Ich|Aber|Ja|Frau|Herr+EIG:/isPosTagUnknown;
// toAddedLemma SUB:. DEFAULT_MIN_PERCENT=5; MIN_WORD_COUNT=100 (abstract); default off.
type StyleTooOftenUsedNounRule struct {
	*rules.AbstractStyleTooOftenUsedWordRule
}

// styleTooOftenNounConfigureText is Java StyleTooOftenUsedNounRule.getConfigureText.
const styleTooOftenNounConfigureText = "Anzeigen wenn ein Substantiv häufiger verwendet wird als ...% aller Substantive:"

// styleTooOftenDefaultMinPercent is Java StyleTooOftenUsedNounRule.DEFAULT_MIN_PERCENT
// (shared constant name with adjective/verb DE style rules: all use 5).
const styleTooOftenDefaultMinPercent = 5

// NewStyleTooOftenUsedNounRule ports
// StyleTooOftenUsedNounRule(messages, lang, null).
func NewStyleTooOftenUsedNounRule(messages map[string]string) *StyleTooOftenUsedNounRule {
	return NewStyleTooOftenUsedNounRuleWithUserConfig(messages, nil)
}

// NewStyleTooOftenUsedNounRuleWithUserConfig ports
// StyleTooOftenUsedNounRule(ResourceBundle messages, Language lang, UserConfig userConfig).
// Java: super(messages, lang, userConfig, DEFAULT_MIN_PERCENT); then if userConfig!=null and
// linguServices!=null → linguServices.setThesaurusRelevantRule(this).
// UserConfig minPercent: getConfigValueByID(getId())[0] when present (abstract getMinPercent).
func NewStyleTooOftenUsedNounRuleWithUserConfig(messages map[string]string, userConfig *languagetool.UserConfig) *StyleTooOftenUsedNounRule {
	minPercent := styleTooOftenDefaultMinPercent
	if userConfig != nil {
		if cf := userConfig.GetConfigValueByID("TOO_OFTEN_USED_NOUN_DE"); len(cf) > 0 {
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
		ID:                "TOO_OFTEN_USED_NOUN_DE",
		Description:       "Statistische Stilanalyse: Zu häufig genutztes Substantiv",
		ConfigureText:     styleTooOftenNounConfigureText,
		DefaultMinPercent: styleTooOftenDefaultMinPercent,
		MinPercent:        minPercent,
		MinWordCount:      100, // Java AbstractStyleTooOftenUsedWordRule.MIN_WORD_COUNT
		// Java getLimitMessage — no space before period (unlike StyleTooOftenUsedAdjectiveRule).
		LimitMessage: func(limit int) string {
			return "Das Substantiv wird häufiger verwendet als " + itoaDE(limit) +
				"% aller Substantive. Möglicherweise ist es besser es durch ein Synonym zu ersetzen."
		},
		// Java: isToCountedWord = hasPosTagStartingWith("SUB:")
		IsToCountedWord: func(tok *languagetool.AnalyzedTokenReadings) bool {
			return tok != nil && tok.HasPosTagStartingWith("SUB:")
		},
		// Java isException:
		//   token.hasPosTagStartingWith("PRO:")
		//   || token.getToken().equals("Ich") || "Aber" || "Ja"
		//   || (n < tokens.length-1 && (Frau|Herr)
		//       && (tokens[n+1].hasPosTagStartingWith("EIG:") || tokens[n+1].isPosTagUnknown()))
		// Critical: isPosTagUnknown() ≠ !isTagged() — Java flag is set only when constructed
		// with a single reading whose POS is null (AnalyzedTokenReadings ctor).
		IsException: func(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
			token := tokens[n]
			if token.HasPosTagStartingWith("PRO:") {
				return true
			}
			switch token.GetToken() {
			case "Ich", "Aber", "Ja":
				return true
			}
			if n < len(tokens)-1 && tokens[n+1] != nil &&
				(token.GetToken() == "Frau" || token.GetToken() == "Herr") &&
				(tokens[n+1].HasPosTagStartingWith("EIG:") || tokens[n+1].IsPosTagUnknown()) {
				return true
			}
			return false
		},
		// Java: toAddedLemma = getLemmaForPosTagStartsWith("SUB:", token)
		ToAddedLemma: func(tok *languagetool.AnalyzedTokenReadings) string {
			return rules.LemmaForPosTagStartsWith("SUB:", tok)
		},
	}
	rules.InitStyleTooOftenUsedWordMeta(base, messages, false)
	r := &StyleTooOftenUsedNounRule{AbstractStyleTooOftenUsedWordRule: base}
	// Java ctor: if (userConfig != null) { LinguServices ls = userConfig.getLinguServices();
	// if (ls != null) ls.setThesaurusRelevantRule(this); }
	if userConfig != nil {
		if ls := userConfig.GetLinguServices(); ls != nil {
			ls.SetThesaurusRelevantRule(r)
		}
	}
	return r
}

// NewStyleTooOftenUsedNounRuleWithDefaultLimit is the register-core entry (Java defaults).
func NewStyleTooOftenUsedNounRuleWithDefaultLimit(messages map[string]string) *StyleTooOftenUsedNounRule {
	r := NewStyleTooOftenUsedNounRule(messages)
	r.MinPercent = styleTooOftenDefaultMinPercent
	r.MinWordCount = 100
	return r
}

func (r *StyleTooOftenUsedNounRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractStyleTooOftenUsedWordRule.MatchList(sentences)
}
