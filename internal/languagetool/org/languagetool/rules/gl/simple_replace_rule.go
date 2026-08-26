package gl

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/words.txt
var wordsFS embed.FS

// SimpleReplaceRuleID ports SimpleReplaceRule.GL_SIMPLE_REPLACE_RULE → "GL_SIMPLE_REPLACE".
const SimpleReplaceRuleID = "GL_SIMPLE_REPLACE"

// SimpleReplaceFilePath is the official classpath resource Java loads
// (static loadFromPath("/gl/words.txt")).
const SimpleReplaceFilePath = "/gl/words.txt"

var (
	wordsOnce sync.Once
	wordsMap  map[string][]string
)

// loadWords ports static wrongWords = loadFromPath("/gl/words.txt").
// Official resource is go:embed of the same words.txt Java loads (byte-identical).
func loadWords() map[string][]string {
	wordsOnce.Do(func() {
		f, err := wordsFS.Open("data/words.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		wordsMap = m
	})
	return wordsMap
}

// SimpleReplaceRule ports org.languagetool.rules.gl.SimpleReplaceRule
// (extends AbstractSimpleReplaceRule, not AbstractSimpleReplaceRule2).
//
// Java field twin summary:
//   - id GL_SIMPLE_REPLACE
//   - static wrongWords from /gl/words.txt
//   - Categories.MISC via AbstractSimpleReplaceRule super (no setCategory override)
//   - no setLocQualityIssueType → base ITSIssueType default (Uncategorized)
//   - isCaseSensitive() false
//   - isCheckLemmas default true (ASR field; SimpleReplaceRule does not call setCheckLemmas(false))
//   - getLocale() → Locale("gl")
//   - description / short / message hardcoded Galician strings
//   - no addExamplePair, no setIgnoreTaggedWords, no useSubRuleSpecificIds, no setUrl, no setDefaultOff
type SimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewSimpleReplaceRule ports SimpleReplaceRule(ResourceBundle messages, Language language).
// language is represented via LanguageCode "gl" (getLocale → Locale("gl")).
func NewSimpleReplaceRule(messages map[string]string) *SimpleReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadWords(),
		CaseSensitive: false, // Java isCaseSensitive() false
		CheckLemmas:   true,  // Java AbstractSimpleReplaceRule default true
		ID:            SimpleReplaceRuleID,
		LanguageCode:  "gl", // Java: new Locale("gl")
		Description:   "Corrección de erros léxicos (barbarismos).",
		ShortMsg:      "Erros léxicos",
		// Java AbstractSimpleReplaceRule ctor: Categories.MISC.getCategory(messages)
		Category: rules.CatMisc.GetCategory(messages),
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java: "'" + tokenStr + "' non existe en galego. Talvez quería vostede dicir: "
			//       + String.join(", ", replacements) + "."
			return "'" + tokenStr + "' non existe en galego. Talvez quería vostede dicir: " +
				strings.Join(replacements, ", ") + "."
		},
	}
	return &SimpleReplaceRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports SimpleReplaceRule.getShort() → "Erros léxicos".
func (r *SimpleReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports SimpleReplaceRule.getMessage(token, replacements)
// with empty token/replacements for the static template surface (tests / metadata).
// Match-time messages go through MessageFn with real suggestions.
func (r *SimpleReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn("", nil)
	}
	return ""
}

// GetLocaleCode ports SimpleReplaceRule.getLocale() → Locale("gl").
func (r *SimpleReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports SimpleReplaceRule.getWrongWords() → static map.
func (r *SimpleReplaceRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// IsDefaultOff ports Rule.isDefaultOff — Java never calls setDefaultOff.
func (r *SimpleReplaceRule) IsDefaultOff() bool {
	return false
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *SimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
