package pt

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace.txt
var ptReplaceFS embed.FS

// PortugueseReplaceID ports PortugueseReplaceRule.PORTUGUESE_SIMPLE_REPLACE_RULE.
const PortugueseReplaceID = "PT_SIMPLE_REPLACE"

// PortugueseReplaceFilePath is the official classpath resource Java loads
// (static loadFromPath("/pt/replace.txt")).
const PortugueseReplaceFilePath = "/pt/replace.txt"

// PortugueseReplaceURL ports PortugueseReplaceRule.getUrl()
// → Tools.getUrl("https://pt.wikipedia.org/wiki/Estrangeirismo").
const PortugueseReplaceURL = "https://pt.wikipedia.org/wiki/Estrangeirismo"

var (
	ptReplaceOnce sync.Once
	ptReplaceMap  map[string][]string
)

// loadPTReplace ports static wrongWords = loadFromPath("/pt/replace.txt").
// Official file is comments-only (no active pairs); empty map is correct, not invent.
func loadPTReplace() map[string][]string {
	ptReplaceOnce.Do(func() {
		f, err := ptReplaceFS.Open("data/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		ptReplaceMap = m
	})
	return ptReplaceMap
}

// PortugueseReplaceRule ports
// org.languagetool.rules.pt.PortugueseReplaceRule
// (extends AbstractSimpleReplaceRule).
//
// Java addExamplePair(device→dispositivo) is commented out — zero examples.
// Wired from base Portuguese language (not only BR/PT variants).
type PortugueseReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewPortugueseReplaceRule ports
// PortugueseReplaceRule(ResourceBundle messages, Language language).
// language is Portuguese (short code "pt") for Locale("pt") / StringTools.toId.
func NewPortugueseReplaceRule(messages map[string]string) *PortugueseReplaceRule {
	// Field twin of PortugueseReplaceRule + AbstractSimpleReplaceRule defaults:
	// - id PT_SIMPLE_REPLACE
	// - static wrongWords from /pt/replace.txt (comments-only → empty map)
	// - isCaseSensitive() false
	// - checkLemmas default true (not setCheckLemmas(false))
	// - getLocale → Locale("pt") via LanguageCode
	// - useSubRuleSpecificIds()
	// - Categories.STYLE, ITSIssueType.LocaleViolation
	// - URL wikipedia Estrangeirismo
	// - description / short / message hardcoded PT strings (not ResourceBundle keys)
	// - example pair COMMENTED OUT in Java → no AddExamplePair
	base := &rules.AbstractSimpleReplaceRule{
		Messages:           messages,
		WrongWords:         loadPTReplace(),
		CaseSensitive:      false,
		CheckLemmas:        true, // Java ASR1 default; not setCheckLemmas(false)
		ID:                 PortugueseReplaceID,
		LanguageCode:       "pt",
		SubRuleSpecificIDs: true,
		Description:        "Palavras estrangeiras facilmente confundidas em Português",
		ShortMsg:           "Estrangeirismo",
		Category:           rules.CatStyle.GetCategory(messages),
		IssueType:          rules.ITSLocaleViolation,
		URL:                PortugueseReplaceURL,
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java getMessage: "'" + tokenStr + "' é um estrangeirismo. Em Português é mais comum usar: "
			//   + String.join(", ", replacements) + "."
			return "'" + tokenStr + "' é um estrangeirismo. Em Português é mais comum usar: " +
				strings.Join(replacements, ", ") + "."
		},
	}
	// Java example pair is commented out — do not invent live Example.
	return &PortugueseReplaceRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports PortugueseReplaceRule.getShort().
func (r *PortugueseReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports PortugueseReplaceRule.getMessage(token, replacements)
// with empty token/replacements for the static template surface (tests / metadata).
// Match-time messages go through MessageFn with real token and suggestions.
func (r *PortugueseReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn("", nil)
	}
	return ""
}

// GetLocaleCode ports PortugueseReplaceRule.getLocale() → Locale("pt").
func (r *PortugueseReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports PortugueseReplaceRule.getWrongWords() → static map.
func (r *PortugueseReplaceRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *PortugueseReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
