package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace.txt data/replace_custom.txt
var enReplaceFS embed.FS

// EnglishSimpleReplaceRuleID ports SimpleReplaceRule.getId() → "EN_SIMPLE_REPLACE".
const EnglishSimpleReplaceRuleID = "EN_SIMPLE_REPLACE"

// EnglishReplaceFilePath ports SimpleReplaceRule.getFileNames()[0] → "/en/replace.txt".
const EnglishReplaceFilePath = "/en/replace.txt"

// EnglishReplaceCustomFilePath ports SimpleReplaceRule.getFileNames()[1] → "/en/replace_custom.txt".
const EnglishReplaceCustomFilePath = "/en/replace_custom.txt"

// EnglishSimpleReplaceFileNames ports SimpleReplaceRule.getFileNames().
var EnglishSimpleReplaceFileNames = []string{EnglishReplaceFilePath, EnglishReplaceCustomFilePath}

// englishSimpleReplaceEmbedPaths pairs go:embed paths with Java rules-dir paths
// (AbstractSimpleReplaceRule2.getAllFilePaths → getFromRulesDirAsUrl).
var englishSimpleReplaceEmbedPaths = []struct {
	embed string
	java  string
}{
	{"data/replace.txt", EnglishReplaceFilePath},
	{"data/replace_custom.txt", EnglishReplaceCustomFilePath},
}

var (
	enReplaceOnce sync.Once
	enReplaceBase *rules.AbstractSimpleReplaceRule2
)

// loadENReplace loads official replace.txt + replace_custom.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit / fillMaps over getFileNames()).
func loadENReplace() *rules.AbstractSimpleReplaceRule2 {
	enReplaceOnce.Do(func() {
		// Field twin of SimpleReplaceRule + AbstractSimpleReplaceRule2 defaults:
		// - id EN_SIMPLE_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator default ", " (Java getSuggestionsSeparator not overridden)
		// - CaseSensitivy.CI (base default)
		// - Locale EN via LanguageCode (getLocale → new Locale("EN"))
		// - useSubRuleSpecificIds()
		// - Categories.MISC (AbstractSimpleReplaceRule2 ctor)
		// - ITSIssueType.Uncategorized (Rule default; not overridden)
		// - no example pairs in Java ctor
		base := &rules.AbstractSimpleReplaceRule2{
			ID:              EnglishSimpleReplaceRuleID,
			Description:     "Check for wrong words/phrases: $match",
			ShortMsg:        "Wrong word",
			MessageTemplate: "Did you mean $suggestions?",
			// SuggestionsSeparator left empty → Match uses Java default ", "
			CaseSens:           rules.CaseInsensitive,
			LanguageCode:       "EN",
			SubRuleSpecificIDs: true,
			Category:           rules.CatMisc.GetCategory(nil),
		}
		for _, p := range englishSimpleReplaceEmbedPaths {
			f, err := enReplaceFS.Open(p.embed)
			if err != nil {
				// Java getFromRulesDirAsUrl / openStream fails closed if resource missing.
				panic(err)
			}
			if err := base.LoadSimpleReplaceRule2Data(f, p.java); err != nil {
				f.Close()
				panic(err)
			}
			f.Close()
		}
		enReplaceBase = base
	})
	return enReplaceBase
}

// SimpleReplaceRule ports org.languagetool.rules.en.SimpleReplaceRule
// (extends AbstractSimpleReplaceRule2).
type SimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewSimpleReplaceRule ports SimpleReplaceRule(ResourceBundle messages, Language language).
// Official data is go:embed of the same replace.txt / replace_custom.txt Java loads.
// language is English (short code "en"); getLocale is Locale("EN") — stored as LanguageCode.
func NewSimpleReplaceRule(messages map[string]string) *SimpleReplaceRule {
	base := loadENReplace()
	r := *base
	r.Messages = messages
	r.Category = rules.CatMisc.GetCategory(messages)
	return &SimpleReplaceRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports SimpleReplaceRule.getFileNames().
func (r *SimpleReplaceRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{EnglishReplaceFilePath, EnglishReplaceCustomFilePath}
}

// GetShort ports SimpleReplaceRule.getShort().
func (r *SimpleReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports SimpleReplaceRule.getMessage().
func (r *SimpleReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetLocaleCode ports SimpleReplaceRule.getLocale() → Locale("EN").
func (r *SimpleReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from replace*.txt).
func (r *SimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
