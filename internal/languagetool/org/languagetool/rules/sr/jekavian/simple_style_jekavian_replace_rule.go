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
//	load("/sr/jekavian/replace-style.txt")
//
//go:embed data/replace-style.txt
var styleFS embed.FS

// SimpleStyleJekavianReplaceRuleID ports SimpleStyleJekavianReplaceRule.getId().
const SimpleStyleJekavianReplaceRuleID = "SR_JEKAVIAN_SIMPLE_STYLE_REPLACE_RULE"

// SimpleStyleJekavianReplaceStylePath ports the resource path loaded by Java.
const SimpleStyleJekavianReplaceStylePath = "/sr/jekavian/replace-style.txt"

var (
	styleOnce sync.Once
	styleMap  map[string][]string
)

func loadStyle() map[string][]string {
	styleOnce.Do(func() {
		f, err := styleFS.Open("data/replace-style.txt")
		if err != nil {
			// Java load fails closed if resource missing.
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		styleMap = m
	})
	return styleMap
}

// SimpleStyleJekavianReplaceRule ports
// org.languagetool.rules.sr.jekavian.SimpleStyleJekavianReplaceRule
// (extends AbstractSimpleReplaceRule). Loads fixed /sr/jekavian/replace-style.txt.
// Official resource is header-only (comment lines only) → empty wrongWords map.
type SimpleStyleJekavianReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewSimpleStyleJekavianReplaceRule ports
// SimpleStyleJekavianReplaceRule(ResourceBundle messages).
// Java: super(messages) → AbstractSimpleReplaceRule sets Categories.MISC;
// checkLemmas default true; isCaseSensitive default true (not overridden);
// getLocale → Locale("sr").
func NewSimpleStyleJekavianReplaceRule(messages map[string]string) *SimpleStyleJekavianReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:   messages,
		WrongWords: loadStyle(),
		// Java AbstractSimpleReplaceRule.isCaseSensitive() default true;
		// SimpleStyleJekavianReplaceRule does not override.
		CaseSensitive: true,
		// Java AbstractSimpleReplaceRule checkLemmas field default true.
		CheckLemmas: true,
		ID:          SimpleStyleJekavianReplaceRuleID,
		// LanguageCode stores getLocale() → Locale("sr") for case-conversion channel.
		LanguageCode: "sr",
		Description:  "Провера стилски лоших ријечи или израза",
		ShortMsg:     "Стилски лоша ријеч тј. израз",
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java getMessage:
			// "Умјесто израза „" + tokenStr + "“ било би боље да користите: "
			//   + String.join(", ", replacements) + "."
			// Quotation marks: U+201E „ and U+201C “.
			return "Умјесто израза „" + tokenStr + "“ било би боље да користите: " +
				strings.Join(replacements, ", ") + "."
		},
	}
	// Java AbstractSimpleReplaceRule ctor: Categories.MISC.getCategory(messages).
	rules.InitSimpleReplaceMeta(base, messages)
	return &SimpleStyleJekavianReplaceRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports SimpleStyleJekavianReplaceRule.getShort().
func (r *SimpleStyleJekavianReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports SimpleStyleJekavianReplaceRule.getMessage(tokenStr, replacements).
func (r *SimpleStyleJekavianReplaceRule) GetMessage(tokenStr string, replacements []string) string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn(tokenStr, replacements)
	}
	return ""
}

// GetLocaleCode ports SimpleStyleJekavianReplaceRule.getLocale() → Locale("sr").
func (r *SimpleStyleJekavianReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports SimpleStyleJekavianReplaceRule.getWrongWords().
func (r *SimpleStyleJekavianReplaceRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *SimpleStyleJekavianReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
