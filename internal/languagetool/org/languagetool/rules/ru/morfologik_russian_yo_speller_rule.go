package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	// MorfologikRussianYOSpellerRuleID ports MorfologikRussianYOSpellerRule.RULE_ID / getId().
	// Java: "MORFOLOGIK_RULE_RU_RU_YO" (not MORFOLOGIK_RULE_RU_YO).
	MorfologikRussianYOSpellerRuleID = "MORFOLOGIK_RULE_RU_RU_YO"
	// RussianYOSpellerDict ports MorfologikRussianYOSpellerRule.RESOURCE_FILENAME / getFileName().
	// Java: "/ru/hunspell/ru_RU_yo.dict"
	RussianYOSpellerDict = "/ru/hunspell/ru_RU_yo.dict"
)

// Java YO-only extra NOSUGGEST (елка + shared set on base).
var ruYOExtraDoNotSuggest = map[string]struct{}{
	"елка": {},
}

// MorfologikRussianYOSpellerRule ports rules.ru.MorfologikRussianYOSpellerRule
// (ё-aware dict). Default off in RegisterCore (Java setDefaultOff).
type MorfologikRussianYOSpellerRule struct {
	*MorfologikRussianSpellerRule
}

func NewMorfologikRussianYOSpellerRule() *MorfologikRussianYOSpellerRule {
	base := &MorfologikRussianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikRussianYOSpellerRuleID, "ru", RussianYOSpellerDict, nil),
		ConfCheckLatin:    defaultMinRUValue,
		ExtraDoNotSuggest: ruYOExtraDoNotSuggest,
	}
	// Java isLatinScript() = false
	if base.SpellingCheckRule != nil {
		base.NonLatinScript = true
		base.FilterNoSuggestWordsFn = base.filterNoSuggestWords
	}
	base.IgnoreTokenFn = base.ruIgnoreToken
	// Java YO ctor: same demos as base speller (каждя → каждая)
	base.AddExamplePair(
		rules.Wrong("Все счастливые семьи похожи друг на друга, <marker>каждя</marker> несчастливая семья несчастлива по-своему."),
		rules.Fixed("Все счастливые семьи похожи друг на друга, <marker>каждая</marker> несчастливая семья несчастлива по-своему."),
	)
	return &MorfologikRussianYOSpellerRule{MorfologikRussianSpellerRule: base}
}

// GetDescription ports getDescription (ё-only experimental rule).
func (r *MorfologikRussianYOSpellerRule) GetDescription() string {
	return "Проверка орфографии. Только «Ё» (экспериментальное правило)."
}
