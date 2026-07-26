package ru

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official replace list Java loads via RussianSimpleReplaceRule:
//
//	getFileNames() → singleton "/ru/replace.txt"
//
//go:embed data/replace.txt
var replaceFS embed.FS

// RussianSimpleReplaceRuleID ports RussianSimpleReplaceRule.RUSSIAN_SIMPLE_REPLACE_RULE.
const RussianSimpleReplaceRuleID = "RU_SIMPLE_REPLACE"

// RussianReplaceFilePath ports RussianSimpleReplaceRule.getFileNames()[0] → "/ru/replace.txt".
const RussianReplaceFilePath = "/ru/replace.txt"

// RussianSimpleReplaceFileNames ports RussianSimpleReplaceRule.getFileNames().
var RussianSimpleReplaceFileNames = []string{RussianReplaceFilePath}

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
		// Field twin of RussianSimpleReplaceRule + AbstractSimpleReplaceRule2 defaults:
		// - id RU_SIMPLE_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator default ", " (Java getSuggestionsSeparator not overridden)
		// - CaseSensitivy.CI (base default)
		// - Locale "ru" via LanguageCode (getLocale → new Locale("ru"))
		// - does NOT call useSubRuleSpecificIds()
		// - Category(new CategoryId("MISC"), "Общие правила") after super MISC
		// - ITSIssueType.Misspelling
		// - example pair Экспрессо → Эспрессо
		base := &rules.AbstractSimpleReplaceRule2{
			ID:              RussianSimpleReplaceRuleID,
			Description:     "Поиск просторечий и ошибочных фраз",
			ShortMsg:        "Ошибка?",
			MessageTemplate: "«$match» — просторечие, исправление: $suggestions",
			// SuggestionsSeparator left empty → Match uses Java default ", "
			CaseSens:           rules.CaseInsensitive,
			LanguageCode:       "ru",
			SubRuleSpecificIDs: false, // Java never calls useSubRuleSpecificIds()
			Category: rules.NewCategory(
				rules.NewCategoryId("MISC"),
				"Общие правила",
			),
			IssueType: rules.ITSMisspelling,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, RussianReplaceFilePath); err != nil {
			panic(err)
		}
		// Java:
		// addExamplePair(Example.wrong("<marker>Экспрессо</marker> – крепкий кофе, …"),
		//                Example.fixed("<marker>Эспрессо</marker> – крепкий кофе, …"));
		base.AddExamplePair(
			rules.Wrong("<marker>Экспрессо</marker> – крепкий кофе, приготовленный из хорошо обжаренных и тонко помолотых кофейных зёрен."),
			rules.Fixed("<marker>Эспрессо</marker> – крепкий кофе, приготовленный из хорошо обжаренных и тонко помолотых кофейных зёрен."),
		)
		replaceBase = base
	})
	return replaceBase
}

// RussianSimpleReplaceRule ports org.languagetool.rules.ru.RussianSimpleReplaceRule
// (extends AbstractSimpleReplaceRule2). Loads fixed /ru/replace.txt.
type RussianSimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewRussianSimpleReplaceRule ports RussianSimpleReplaceRule(ResourceBundle messages).
// Java: super(messages, Russian.getInstance()); setLocQualityIssueType(Misspelling);
// setCategory(new Category(new CategoryId("MISC"), "Общие правила")); addExamplePair(…).
// language is represented via LanguageCode "ru" on the base (getLocale → Locale("ru")).
func NewRussianSimpleReplaceRule(messages map[string]string) *RussianSimpleReplaceRule {
	base := loadReplace()
	r := *base
	r.Messages = messages
	r.Category = rules.NewCategory(rules.NewCategoryId("MISC"), "Общие правила")
	r.IssueType = rules.ITSMisspelling
	return &RussianSimpleReplaceRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports RussianSimpleReplaceRule.getFileNames()
// → Collections.singletonList("/ru/replace.txt").
func (r *RussianSimpleReplaceRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{RussianReplaceFilePath}
}

// GetShort ports RussianSimpleReplaceRule.getShort().
func (r *RussianSimpleReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports RussianSimpleReplaceRule.getMessage().
func (r *RussianSimpleReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetLocaleCode ports RussianSimpleReplaceRule.getLocale() → Locale("ru").
func (r *RussianSimpleReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from replace.txt).
func (r *RussianSimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
