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
//	load("/sr/ekavian/replace-style.txt")
//
//go:embed data/replace-style.txt
var styleFS embed.FS

// SimpleStyleEkavianReplaceRuleID ports SimpleStyleEkavianReplaceRule.getId().
const SimpleStyleEkavianReplaceRuleID = "SR_EKAVIAN_SIMPLE_STYLE_REPLACE_RULE"

// SimpleStyleEkavianReplaceStylePath ports the resource path loaded by Java.
const SimpleStyleEkavianReplaceStylePath = "/sr/ekavian/replace-style.txt"

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

// SimpleStyleEkavianReplaceRule ports
// org.languagetool.rules.sr.ekavian.SimpleStyleEkavianReplaceRule
// (extends AbstractSimpleReplaceRule). Loads fixed /sr/ekavian/replace-style.txt.
type SimpleStyleEkavianReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewSimpleStyleEkavianReplaceRule ports
// SimpleStyleEkavianReplaceRule(ResourceBundle messages).
// Java: super(messages) → AbstractSimpleReplaceRule sets Categories.MISC;
// checkLemmas default true; isCaseSensitive default true (not overridden);
// getLocale → Locale("sr").
func NewSimpleStyleEkavianReplaceRule(messages map[string]string) *SimpleStyleEkavianReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:   messages,
		WrongWords: loadStyle(),
		// Java AbstractSimpleReplaceRule.isCaseSensitive() default true;
		// SimpleStyleEkavianReplaceRule does not override.
		CaseSensitive: true,
		// Java AbstractSimpleReplaceRule checkLemmas field default true.
		CheckLemmas: true,
		ID:          SimpleStyleEkavianReplaceRuleID,
		// LanguageCode stores getLocale() → Locale("sr") for case-conversion channel.
		LanguageCode: "sr",
		Description:  "Провера стилски лоших речи или израза",
		ShortMsg:     "Стилски лоша реч тј. израз",
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java getMessage:
			// "Уместо израза „" + tokenStr + "“ било би боље да користите: "
			//   + String.join(", ", replacements) + "."
			// Quotation marks: U+201E „ and U+201C “.
			return "Уместо израза „" + tokenStr + "“ било би боље да користите: " +
				strings.Join(replacements, ", ") + "."
		},
	}
	// Java AbstractSimpleReplaceRule ctor: Categories.MISC.getCategory(messages).
	rules.InitSimpleReplaceMeta(base, messages)
	return &SimpleStyleEkavianReplaceRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports SimpleStyleEkavianReplaceRule.getShort().
func (r *SimpleStyleEkavianReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports SimpleStyleEkavianReplaceRule.getMessage(tokenStr, replacements).
func (r *SimpleStyleEkavianReplaceRule) GetMessage(tokenStr string, replacements []string) string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn(tokenStr, replacements)
	}
	return ""
}

// GetLocaleCode ports SimpleStyleEkavianReplaceRule.getLocale() → Locale("sr").
func (r *SimpleStyleEkavianReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports SimpleStyleEkavianReplaceRule.getWrongWords().
func (r *SimpleStyleEkavianReplaceRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *SimpleStyleEkavianReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
