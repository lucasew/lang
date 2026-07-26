package jekavian

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official path Java loads:
//
//	load("/sr/jekavian/replace-grammar.txt")
//
//go:embed data/replace-grammar.txt
var grammarFS embed.FS

// SimpleGrammarJekavianReplaceRuleID ports SimpleGrammarJekavianReplaceRule.getId().
const SimpleGrammarJekavianReplaceRuleID = "SR_JEKAVIAN_SIMPLE_GRAMMAR_REPLACE_RULE"

// SimpleGrammarJekavianReplaceGrammarPath ports the resource path loaded by Java.
const SimpleGrammarJekavianReplaceGrammarPath = "/sr/jekavian/replace-grammar.txt"

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

// SimpleGrammarJekavianReplaceRule ports
// org.languagetool.rules.sr.jekavian.SimpleGrammarJekavianReplaceRule
// (extends AbstractSimpleReplaceRule). Loads fixed /sr/jekavian/replace-grammar.txt.
// Official resource is header-only (comment lines only) → empty wrongWords map.
type SimpleGrammarJekavianReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewSimpleGrammarJekavianReplaceRule ports
// SimpleGrammarJekavianReplaceRule(ResourceBundle messages).
// Java: super(messages) → AbstractSimpleReplaceRule sets Categories.MISC;
// checkLemmas default true; isCaseSensitive default true (not overridden);
// getLocale → Locale("sr").
func NewSimpleGrammarJekavianReplaceRule(messages map[string]string) *SimpleGrammarJekavianReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:   messages,
		WrongWords: loadGrammar(),
		// Java AbstractSimpleReplaceRule.isCaseSensitive() default true;
		// SimpleGrammarJekavianReplaceRule does not override.
		CaseSensitive: true,
		// Java AbstractSimpleReplaceRule checkLemmas field default true.
		CheckLemmas: true,
		ID:          SimpleGrammarJekavianReplaceRuleID,
		// LanguageCode stores getLocale() → Locale("sr") for case-conversion channel.
		LanguageCode: "sr",
		Description:  "Провера граматички погрешних ријечи или израза",
		ShortMsg:     "Граматички погрешна ријеч тј. израз",
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java getMessage:
			// "Не каже се „" + tokenStr + "“ него „" + String.join(", ", replacements) + "“."
			// Quotation marks: U+201E „ and U+201C “.
			return "Не каже се „" + tokenStr + "“ него „" + strings.Join(replacements, ", ") + "“."
		},
	}
	// Java AbstractSimpleReplaceRule ctor: Categories.MISC.getCategory(messages).
	rules.InitSimpleReplaceMeta(base, messages)
	return &SimpleGrammarJekavianReplaceRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports SimpleGrammarJekavianReplaceRule.getShort().
func (r *SimpleGrammarJekavianReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports SimpleGrammarJekavianReplaceRule.getMessage(tokenStr, replacements).
func (r *SimpleGrammarJekavianReplaceRule) GetMessage(tokenStr string, replacements []string) string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn(tokenStr, replacements)
	}
	return ""
}

// GetLocaleCode ports SimpleGrammarJekavianReplaceRule.getLocale() → Locale("sr").
func (r *SimpleGrammarJekavianReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports SimpleGrammarJekavianReplaceRule.getWrongWords().
func (r *SimpleGrammarJekavianReplaceRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *SimpleGrammarJekavianReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}

// joinComma is shared package helper used by sibling style rule in this package.
// (Style sector still references it; grammar rule uses strings.Join per Java String.join.)
func joinComma(ss []string) string {
	return strings.Join(ss, ", ")
}
