package fa

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace.txt
var replaceFS embed.FS

// FASimpleReplaceID ports SimpleReplaceRule.getId() → "FA_SIMPLE_REPLACE".
const FASimpleReplaceID = "FA_SIMPLE_REPLACE"

// FAReplaceFilePath is the official classpath resource Java loads
// (static loadFromPath("/fa/replace.txt")).
const FAReplaceFilePath = "/fa/replace.txt"

var (
	replaceOnce sync.Once
	replaceMap  map[string][]string
)

// loadReplace ports static wrongWords = loadFromPath("/fa/replace.txt").
// Official resource is go:embed of the same replace.txt Java loads (byte-identical).
func loadReplace() map[string][]string {
	replaceOnce.Do(func() {
		f, err := replaceFS.Open("data/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		replaceMap = m
	})
	return replaceMap
}

// SimpleReplaceRule ports org.languagetool.rules.fa.SimpleReplaceRule
// (extends AbstractSimpleReplaceRule, not AbstractSimpleReplaceRule2).
//
// Java field twin summary:
//   - id FA_SIMPLE_REPLACE
//   - static wrongWords from /fa/replace.txt
//   - Categories.CONFUSED_WORDS (overrides ASR ctor MISC)
//   - ITSIssueType.Misspelling
//   - isCaseSensitive() false
//   - isCheckLemmas default true (ASR field; FA does not call setCheckLemmas(false))
//   - getLocale() → Locale.getDefault() (not Locale("fa"))
//   - description / short / message hardcoded Persian strings
//   - example: حاظر → حاضر
//   - no setIgnoreTaggedWords, no useSubRuleSpecificIds, no setUrl
//
// Locale.getDefault() mapping: AbstractSimpleReplaceRule.cleanup uses
// strings.ToLower (Unicode) for !caseSensitive, matching other JVM-default
// ports (e.g. BretonTagger conversionLocale). FA replace.txt is Persian
// script — case-fold is identity for those tokens; no FA-specific locale
// quirks are invented beyond Java.
type SimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewSimpleReplaceRule ports SimpleReplaceRule(ResourceBundle messages, Language language).
// language is unused by FA overrides (getLocale is Locale.getDefault(), not language locale).
func NewSimpleReplaceRule(messages map[string]string) *SimpleReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadReplace(),
		CaseSensitive: false, // Java isCaseSensitive() false
		CheckLemmas:   true,  // Java AbstractSimpleReplaceRule default true
		ID:            FASimpleReplaceID,
		// LanguageCode left empty: Java getLocale() is Locale.getDefault(), not Locale("fa").
		Description: "اشتباه محتمل املائی",
		ShortMsg:    "اشتباه محتمل املائی",
		Category:    rules.CatConfusedWords.GetCategory(messages),
		IssueType:   rules.ITSMisspelling,
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java: "اشتباه محتمل املائی پیداشده: " + String.join("، ", replacements) + "."
			// tokenStr is unused in the FA message (matches Java getMessage signature).
			return "اشتباه محتمل املائی پیداشده: " + strings.Join(replacements, "، ") + "."
		},
	}
	// Java: حاظر → حاضر
	base.AddExamplePair(
		rules.Wrong("وی <marker>حاظر</marker> به همکاری شد."),
		rules.Fixed("وی <marker>حاضر</marker> به همکاری شد."),
	)
	return &SimpleReplaceRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports SimpleReplaceRule.getShort().
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

// GetLocaleCode ports SimpleReplaceRule.getLocale() → Locale.getDefault().
// Empty string documents JVM default (not Locale("fa")). See type comment for
// cleanup/strings.ToLower mapping.
func (r *SimpleReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	// LanguageCode is intentionally empty (JVM default).
	return r.LanguageCode
}

// GetWrongWords ports SimpleReplaceRule.getWrongWords() → static map.
func (r *SimpleReplaceRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *SimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
