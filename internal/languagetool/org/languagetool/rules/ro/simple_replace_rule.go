package ro

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official replace list Java loads via SimpleReplaceRule:
//
//	getFileNames() → singleton "/ro/replace.txt"
//
//go:embed data/replace.txt
var replaceFS embed.FS

// RomanianSimpleReplaceRuleID ports SimpleReplaceRule.ROMANIAN_SIMPLE_REPLACE_RULE.
const RomanianSimpleReplaceRuleID = "RO_SIMPLE_REPLACE"

// RomanianReplaceFilePath ports SimpleReplaceRule.getFileNames()[0] → "/ro/replace.txt".
const RomanianReplaceFilePath = "/ro/replace.txt"

// RomanianSimpleReplaceFileNames ports SimpleReplaceRule.getFileNames().
var RomanianSimpleReplaceFileNames = []string{RomanianReplaceFilePath}

var (
	replaceOnce sync.Once
	replaceBase *rules.AbstractSimpleReplaceRule2
)

// loadReplaceRule loads official replace.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadReplaceRule() *rules.AbstractSimpleReplaceRule2 {
	replaceOnce.Do(func() {
		f, err := replaceFS.Open("data/replace.txt")
		if err != nil {
			// Java getFromRulesDirAsUrl / openStream fails closed if resource missing.
			panic(err)
		}
		defer f.Close()
		// Field twin of SimpleReplaceRule + AbstractSimpleReplaceRule2 defaults:
		// - id RO_SIMPLE_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator " sau " (Java getSuggestionsSeparator override)
		// - CaseSensitivy.CI (base default; RO does not override)
		// - Locale "ro" via LanguageCode (getLocale → new Locale("ro"))
		// - does NOT call useSubRuleSpecificIds()
		// - Category Categories.MISC (set in ctor + NewSimpleReplaceRule)
		// - no setLocQualityIssueType (default Uncategorized)
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   RomanianSimpleReplaceRuleID,
			Description:          "Cuvinte sau grupuri de cuvinte incorecte sau ieșite din uz",
			ShortMsg:             "Cuvânt incorect sau ieșit din uz",
			MessageTemplate:      "'$match' este incorect sau ieșit din uz, folosiți $suggestions",
			SuggestionsSeparator: " sau ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ro",
			SubRuleSpecificIDs:   false, // Java never calls useSubRuleSpecificIds()
			Category:             rules.CatMisc.GetCategory(nil),
		}
		if err := base.LoadSimpleReplaceRule2Data(f, RomanianReplaceFilePath); err != nil {
			panic(err)
		}
		replaceBase = base
	})
	return replaceBase
}

// SimpleReplaceRule ports org.languagetool.rules.ro.SimpleReplaceRule
// (extends AbstractSimpleReplaceRule2). Loads fixed /ro/replace.txt.
type SimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewSimpleReplaceRule ports SimpleReplaceRule(ResourceBundle messages).
// Java: super(messages, new Romanian()); setCategory(Categories.MISC.getCategory(messages));
// language is represented via LanguageCode "ro" on the base (getLocale → Locale("ro")).
func NewSimpleReplaceRule(messages map[string]string) *SimpleReplaceRule {
	base := loadReplaceRule()
	r := *base
	r.Messages = messages
	// Java SimpleReplaceRule ctor: setCategory(Categories.MISC.getCategory(messages));
	// (AbstractSimpleReplaceRule2 super also sets MISC from messages.)
	r.Category = rules.CatMisc.GetCategory(messages)
	return &SimpleReplaceRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports SimpleReplaceRule.getFileNames()
// → Collections.singletonList("/ro/replace.txt").
func (r *SimpleReplaceRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{RomanianReplaceFilePath}
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

// GetSuggestionsSeparator ports SimpleReplaceRule.getSuggestionsSeparator() → " sau ".
func (r *SimpleReplaceRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	if r.SuggestionsSeparator == "" {
		return ", "
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports SimpleReplaceRule.getLocale() → Locale("ro").
func (r *SimpleReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from replace.txt).
func (r *SimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
