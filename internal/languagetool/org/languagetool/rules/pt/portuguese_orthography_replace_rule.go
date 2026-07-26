package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace_orthography.txt
var orthographyFS embed.FS

// PortugueseSimpleReplaceOrthographyID ports
// PortugueseOrthographyReplaceRule.PORTUGUESE_SIMPLE_REPLACE_ORTHOGRAPHY_RULE.
const PortugueseSimpleReplaceOrthographyID = "PT_SIMPLE_REPLACE_ORTHOGRAPHY"

// PortugueseOrthographyFilePath is the official classpath resource Java loads
// (static loadFromPath("/pt/replace_orthography.txt")).
const PortugueseOrthographyFilePath = "/pt/replace_orthography.txt"

// Message keys from MessagesBundle (Java getDescription/getShort/getMessage).
// TestTools.getMessages("pt") resolves Locale("pt"); there is no MessagesBundle_pt
// (only pt_PT / pt_BR), so ResourceBundle falls back to MessagesBundle.properties
// (English). Production JLanguageTool(Portuguese variant) passes the variant bundle.
const (
	orthographyMsgDescSpelling      = "desc_spelling"
	orthographyMsgDescSpellingShort = "desc_spelling_short"
	orthographyMsgSpelling          = "spelling"
)

// English MessagesBundle.properties surfaces for those keys (Java fallback for Locale("pt")).
const (
	orthographyDefaultDescSpelling      = "Possible spelling mistake"
	orthographyDefaultDescSpellingShort = "Spelling mistake"
	orthographyDefaultSpelling          = "Possible spelling mistake found."
)

var (
	orthographyOnce  sync.Once
	orthographyWords map[string][]string
)

// loadOrthographyWords ports static wrongWords = loadFromPath("/pt/replace_orthography.txt").
func loadOrthographyWords() map[string][]string {
	orthographyOnce.Do(func() {
		f, err := orthographyFS.Open("data/replace_orthography.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		orthographyWords = m
	})
	return orthographyWords
}

// orthographyMsg ports messages.getString(key) with MessagesBundle.properties fallback
// when the map is nil or missing the key (same surface as TestTools.getMessages("pt")).
func orthographyMsg(messages map[string]string, key, fallback string) string {
	if messages != nil {
		if s, ok := messages[key]; ok && s != "" {
			return s
		}
	}
	return fallback
}

// PortugueseOrthographyReplaceRule ports
// org.languagetool.rules.pt.PortugueseOrthographyReplaceRule
// (extends AbstractSimpleReplaceRule).
//
// Multiword immunization (e.g. Italian "sotto voce" in /pt/multiwords.txt) is owned by
// PortugueseHybridDisambiguator MultiWordChunker (setIgnoreSpelling(true)); this rule
// only skips IsImmunized / IsIgnoredBySpeller like Java AbstractSimpleReplaceRule.
// Without that analysis stage, "voce" still matches (fail closed — no surface invent skip).
type PortugueseOrthographyReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewPortugueseOrthographyReplaceRule ports
// PortugueseOrthographyReplaceRule(ResourceBundle messages, Language language).
// language is Portuguese (short code "pt") for Locale("pt") / StringTools.toId.
func NewPortugueseOrthographyReplaceRule(messages map[string]string) *PortugueseOrthographyReplaceRule {
	desc := orthographyMsg(messages, orthographyMsgDescSpelling, orthographyDefaultDescSpelling)
	short := orthographyMsg(messages, orthographyMsgDescSpellingShort, orthographyDefaultDescSpellingShort)
	spelling := orthographyMsg(messages, orthographyMsgSpelling, orthographyDefaultSpelling)
	// Field twin of PortugueseOrthographyReplaceRule + AbstractSimpleReplaceRule defaults:
	// - id PT_SIMPLE_REPLACE_ORTHOGRAPHY
	// - static wrongWords from /pt/replace_orthography.txt
	// - isCaseSensitive() false
	// - checkLemmas default true (not overridden)
	// - getLocale → Locale("pt") via LanguageCode
	// - useSubRuleSpecificIds()
	// - Categories.TYPOS, ITSIssueType.Misspelling
	// - getDescription/getShort/getMessage → ResourceBundle keys (not hardcoded invent)
	base := &rules.AbstractSimpleReplaceRule{
		Messages:           messages,
		WrongWords:         loadOrthographyWords(),
		CaseSensitive:      false,
		CheckLemmas:        true, // Java default; not setCheckLemmas(false)
		ID:                 PortugueseSimpleReplaceOrthographyID,
		LanguageCode:       "pt",
		SubRuleSpecificIDs: true,
		Description:        desc,
		ShortMsg:           short,
		Category:           rules.CatTypos.GetCategory(messages),
		IssueType:          rules.ITSMisspelling,
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java getMessage ignores token/replacements → messages.getString("spelling")
			return spelling
		},
	}
	return &PortugueseOrthographyReplaceRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports PortugueseOrthographyReplaceRule.getShort() → messages "desc_spelling_short".
func (r *PortugueseOrthographyReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports PortugueseOrthographyReplaceRule.getMessage(token, replacements)
// → messages.getString("spelling") (token/replacements unused).
func (r *PortugueseOrthographyReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn("", nil)
	}
	return orthographyDefaultSpelling
}

// GetLocaleCode ports PortugueseOrthographyReplaceRule.getLocale() → Locale("pt").
func (r *PortugueseOrthographyReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports PortugueseOrthographyReplaceRule.getWrongWords() → static map.
func (r *PortugueseOrthographyReplaceRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *PortugueseOrthographyReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
