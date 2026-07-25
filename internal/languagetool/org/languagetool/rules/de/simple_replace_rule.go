package de

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace.txt data/replace_custom.txt
var deReplaceFS embed.FS

// GermanSimpleReplaceRuleID ports SimpleReplaceRule.getId() → "DE_SIMPLE_REPLACE".
const GermanSimpleReplaceRuleID = "DE_SIMPLE_REPLACE"

// GermanReplaceFilePath ports SimpleReplaceRule.getFileNames()[0] → "/de/replace.txt".
const GermanReplaceFilePath = "/de/replace.txt"

// GermanReplaceCustomFilePath ports SimpleReplaceRule.getFileNames()[1] → "/de/replace_custom.txt".
const GermanReplaceCustomFilePath = "/de/replace_custom.txt"

// GermanSimpleReplaceFileNames ports SimpleReplaceRule.getFileNames().
var GermanSimpleReplaceFileNames = []string{GermanReplaceFilePath, GermanReplaceCustomFilePath}

// germanSimpleReplaceEmbedPaths pairs go:embed paths with Java rules-dir paths
// (AbstractSimpleReplaceRule2.getAllFilePaths → getFromRulesDirAsUrl).
var germanSimpleReplaceEmbedPaths = []struct {
	embed string
	java  string
}{
	{"data/replace.txt", GermanReplaceFilePath},
	{"data/replace_custom.txt", GermanReplaceCustomFilePath},
}

var (
	deReplaceOnce sync.Once
	deReplaceBase *rules.AbstractSimpleReplaceRule2
)

// loadDEReplace loads official replace.txt + replace_custom.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit / fillMaps over getFileNames()).
func loadDEReplace() *rules.AbstractSimpleReplaceRule2 {
	deReplaceOnce.Do(func() {
		// Field twin of SimpleReplaceRule + AbstractSimpleReplaceRule2 defaults:
		// - id DE_SIMPLE_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator default ", " (Java getSuggestionsSeparator not overridden)
		// - CaseSensitivy.CI (base default)
		// - Locale DE via LanguageCode (getLocale → new Locale("DE"))
		// - useSubRuleSpecificIds()
		// - Categories.MISC (AbstractSimpleReplaceRule2 ctor)
		// - ITSIssueType.Uncategorized (Rule default; not overridden)
		// - no example pairs in Java ctor
		base := &rules.AbstractSimpleReplaceRule2{
			ID:          GermanSimpleReplaceRuleID,
			Description: "Prüft auf bestimmte falsche Wörter/Phrasen: $match",
			ShortMsg:    "Falsches Wort",
			MessageTemplate: "Meinten Sie vielleicht $suggestions?",
			// SuggestionsSeparator left empty → Match uses Java default ", "
			CaseSens:           rules.CaseInsensitive,
			LanguageCode:       "DE",
			SubRuleSpecificIDs: true,
			Category:           rules.CatMisc.GetCategory(nil),
		}
		for _, p := range germanSimpleReplaceEmbedPaths {
			f, err := deReplaceFS.Open(p.embed)
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
		deReplaceBase = base
	})
	return deReplaceBase
}

// SimpleReplaceRule ports org.languagetool.rules.de.SimpleReplaceRule
// (extends AbstractSimpleReplaceRule2).
type SimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewSimpleReplaceRule ports SimpleReplaceRule(ResourceBundle messages, Language language).
// Official data is go:embed of the same replace.txt / replace_custom.txt Java loads.
// language is German (short code "de"); getLocale is Locale("DE") — stored as LanguageCode.
func NewSimpleReplaceRule(messages map[string]string) *SimpleReplaceRule {
	base := loadDEReplace()
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
	return []string{GermanReplaceFilePath, GermanReplaceCustomFilePath}
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

// GetLocaleCode ports SimpleReplaceRule.getLocale() → Locale("DE").
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
