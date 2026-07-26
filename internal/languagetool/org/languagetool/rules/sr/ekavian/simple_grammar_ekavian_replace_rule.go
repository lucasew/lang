package ekavian

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official path Java loads:
//
//	load("/sr/ekavian/replace-grammar.txt")
//
//go:embed data/replace-grammar.txt
var grammarFS embed.FS

// SimpleGrammarEkavianReplaceRuleID ports SimpleGrammarEkavianReplaceRule.getId().
const SimpleGrammarEkavianReplaceRuleID = "SR_EKAVIAN_SIMPLE_GRAMMAR_REPLACE_RULE"

// SimpleGrammarEkavianReplaceGrammarPath ports the resource path loaded by Java.
const SimpleGrammarEkavianReplaceGrammarPath = "/sr/ekavian/replace-grammar.txt"

var (
	grammarOnce sync.Once
	grammarMap  map[string][]string
)

func loadGrammar() map[string][]string {
	grammarOnce.Do(func() {
		f, err := grammarFS.Open("data/replace-grammar.txt")
		if err != nil {
			// Java load fails closed if resource missing.
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		grammarMap = m
	})
	return grammarMap
}

// SimpleGrammarEkavianReplaceRule ports
// org.languagetool.rules.sr.ekavian.SimpleGrammarEkavianReplaceRule
// (extends AbstractSimpleReplaceRule). Loads fixed /sr/ekavian/replace-grammar.txt.
type SimpleGrammarEkavianReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewSimpleGrammarEkavianReplaceRule ports
// SimpleGrammarEkavianReplaceRule(ResourceBundle messages).
// Java: super(messages) → AbstractSimpleReplaceRule sets Categories.MISC;
// checkLemmas default true; isCaseSensitive default true (not overridden);
// getLocale → Locale("sr").
func NewSimpleGrammarEkavianReplaceRule(messages map[string]string) *SimpleGrammarEkavianReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadGrammar(),
		// Java AbstractSimpleReplaceRule.isCaseSensitive() default true;
		// SimpleGrammarEkavianReplaceRule does not override.
		CaseSensitive: true,
		// Java AbstractSimpleReplaceRule checkLemmas field default true.
		CheckLemmas: true,
		ID:          SimpleGrammarEkavianReplaceRuleID,
		// LanguageCode stores getLocale() → Locale("sr") for case-conversion channel.
		LanguageCode: "sr",
		Description:  "Провера граматички погрешних речи или израза",
		ShortMsg:     "Граматички погрешна реч тј. израз",
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java getMessage:
			// "Не каже се „" + tokenStr + "“ него „" + String.join(", ", replacements) + "“."
			// Quotation marks: U+201E „ and U+201C “.
			return "Не каже се „" + tokenStr + "“ него „" + strings.Join(replacements, ", ") + "“."
		},
	}
	// Java AbstractSimpleReplaceRule ctor: Categories.MISC.getCategory(messages).
	rules.InitSimpleReplaceMeta(base, messages)
	return &SimpleGrammarEkavianReplaceRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports SimpleGrammarEkavianReplaceRule.getShort().
func (r *SimpleGrammarEkavianReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports SimpleGrammarEkavianReplaceRule.getMessage(tokenStr, replacements).
func (r *SimpleGrammarEkavianReplaceRule) GetMessage(tokenStr string, replacements []string) string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn(tokenStr, replacements)
	}
	return ""
}

// GetLocaleCode ports SimpleGrammarEkavianReplaceRule.getLocale() → Locale("sr").
func (r *SimpleGrammarEkavianReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports SimpleGrammarEkavianReplaceRule.getWrongWords().
func (r *SimpleGrammarEkavianReplaceRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *SimpleGrammarEkavianReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
