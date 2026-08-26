package be

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official replace list Java loads via SimpleReplaceRule:
//
//	getFileNames() → singleton "/be/replace.txt"
//
//go:embed data/replace.txt
var replaceFS embed.FS

// BelarusianSimpleReplaceRuleID ports SimpleReplaceRule.BELARUSIAN_SIMPLE_REPLACE_RULE.
const BelarusianSimpleReplaceRuleID = "BE_SIMPLE_REPLACE"

// BelarusianReplaceFilePath ports SimpleReplaceRule.getFileNames()[0] → "/be/replace.txt".
const BelarusianReplaceFilePath = "/be/replace.txt"

// BelarusianSimpleReplaceFileNames ports SimpleReplaceRule.getFileNames().
var BelarusianSimpleReplaceFileNames = []string{BelarusianReplaceFilePath}

var (
	replaceOnce sync.Once
	replaceBase *rules.AbstractSimpleReplaceRule2
)

// loadReplace loads official replace.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadReplace() *rules.AbstractSimpleReplaceRule2 {
	replaceOnce.Do(func() {
		f, err := replaceFS.Open("data/replace.txt")
		if err != nil {
			// Java getFromRulesDirAsUrl / openStream fails closed if resource missing.
			panic(err)
		}
		defer f.Close()
		// Field twin of SimpleReplaceRule + AbstractSimpleReplaceRule2 defaults:
		// - id BE_SIMPLE_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator ", " (Java getSuggestionsSeparator override)
		// - CaseSensitivy.CI (base default)
		// - Locale "be" via LanguageCode (getLocale → new Locale("be"))
		// - does NOT call useSubRuleSpecificIds()
		// - Category(new CategoryId("MISC"), "Агульныя правілы") after super MISC
		// - ITSIssueType.Misspelling
		// - example pair З большага → Збольшага
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   BelarusianSimpleReplaceRuleID,
			Description:          "Пошук прастамоўяў і памылковых фраз",
			ShortMsg:             "Памылка?",
			MessageTemplate:      "«$match» — памылка, нелітаратурны выраз або прастамоўе, правільна: $suggestions",
			SuggestionsSeparator: ", ", // Java getSuggestionsSeparator override
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "be",
			SubRuleSpecificIDs:   false, // Java never calls useSubRuleSpecificIds()
			Category: rules.NewCategory(
				rules.NewCategoryId("MISC"),
				"Агульныя правілы",
			),
			IssueType: rules.ITSMisspelling,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, BelarusianReplaceFilePath); err != nil {
			panic(err)
		}
		// Java:
		// addExamplePair(Example.wrong("<marker>З большага</marker>, гэта быў добры дзень."),
		//                Example.fixed("<marker>Збольшага</marker>, гэта быў добры дзень."));
		base.AddExamplePair(
			rules.Wrong("<marker>З большага</marker>, гэта быў добры дзень."),
			rules.Fixed("<marker>Збольшага</marker>, гэта быў добры дзень."),
		)
		replaceBase = base
	})
	return replaceBase
}

// SimpleReplaceRule ports org.languagetool.rules.be.SimpleReplaceRule
// (extends AbstractSimpleReplaceRule2). Loads fixed /be/replace.txt.
type SimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewSimpleReplaceRule ports SimpleReplaceRule(ResourceBundle messages).
// Java: super(messages, new Belarusian()); setLocQualityIssueType(Misspelling);
// setCategory(new Category(new CategoryId("MISC"), "Агульныя правілы")); addExamplePair(…).
// language is represented via LanguageCode "be" on the base (getLocale → Locale("be")).
func NewSimpleReplaceRule(messages map[string]string) *SimpleReplaceRule {
	base := loadReplace()
	r := *base
	r.Messages = messages
	r.Category = rules.NewCategory(rules.NewCategoryId("MISC"), "Агульныя правілы")
	r.IssueType = rules.ITSMisspelling
	return &SimpleReplaceRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports SimpleReplaceRule.getFileNames()
// → Collections.singletonList("/be/replace.txt").
func (r *SimpleReplaceRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{BelarusianReplaceFilePath}
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

// GetSuggestionsSeparator ports SimpleReplaceRule.getSuggestionsSeparator() → ", ".
func (r *SimpleReplaceRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	if r.SuggestionsSeparator == "" {
		return ", "
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports SimpleReplaceRule.getLocale() → Locale("be").
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
