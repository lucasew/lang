package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	// MorfologikRussianYOSpellerRuleID ports MorfologikRussianYOSpellerRule.RULE_ID / getId().
	// Java: "MORFOLOGIK_RULE_RU_RU_YO"
	MorfologikRussianYOSpellerRuleID = "MORFOLOGIK_RULE_RU_RU_YO"
	// RussianYOSpellerDict ports MorfologikRussianYOSpellerRule.RESOURCE_FILENAME / getFileName().
	// Java: "/ru/hunspell/ru_RU_yo.dict"
	RussianYOSpellerDict = "/ru/hunspell/ru_RU_yo.dict"
)

// Java MorfologikRussianYOSpellerRule.lcDoNotSuggestWords — full set (YO adds "елка" vs base).
// Base ruLcDoNotSuggest covers the shared four; ExtraDoNotSuggest holds the YO-only "елка".
// filterNoSuggestWords merges both (Java: single set with all five).
var ruYOExtraDoNotSuggest = map[string]struct{}{
	"елка": {},
}

// MorfologikRussianYOSpellerRule ports rules.ru.MorfologikRussianYOSpellerRule
// (extends MorfologikSpellerRule; ё-only experimental dict, setDefaultOff).
// Reuses MorfologikRussianSpellerRule for shared ignoreToken / conf / filter surface
// (Java YO and base are near-identical siblings; Go embeds the shared port type).
type MorfologikRussianYOSpellerRule struct {
	*MorfologikRussianSpellerRule
	// DefaultOff ports Rule.setDefaultOff() from Java YO ctor.
	DefaultOff bool
}

// NewMorfologikRussianYOSpellerRule ports ctor with null UserConfig.
func NewMorfologikRussianYOSpellerRule() *MorfologikRussianYOSpellerRule {
	return NewMorfologikRussianYOSpellerRuleWithUser(nil)
}

// NewMorfologikRussianYOSpellerRuleWithUser ports
// MorfologikRussianYOSpellerRule(messages, language, userConfig, altLanguages).
func NewMorfologikRussianYOSpellerRuleWithUser(userConfig *languagetool.UserConfig) *MorfologikRussianYOSpellerRule {
	var userWords []string
	var accepted []string
	if userConfig != nil {
		accepted = userConfig.GetAcceptedWords()
		userWords = morfologik.UserDictWordsForMulti(accepted, userConfig.GetPremiumUid())
	}
	base := &MorfologikRussianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikRussianYOSpellerRuleID, "ru", RussianYOSpellerDict, nil),
		ConfCheckLatin:    defaultMinRUValue,
		ExtraDoNotSuggest: ruYOExtraDoNotSuggest,
	}
	// Java: if (userConfig != null) conf_ru_Value = (int) cf[0] by getId()
	if userConfig != nil {
		if cf := userConfig.GetConfigValueByID(MorfologikRussianYOSpellerRuleID); len(cf) > 0 && cf[0] != nil {
			if n, ok := confRUValueAsInt(cf[0]); ok {
				base.ConfCheckLatin = n
			}
		}
		base.SetUserConfig(userConfig)
	}
	// Java isLatinScript() = false
	if base.SpellingCheckRule != nil {
		base.NonLatinScript = true
		// Java filterNoSuggestWords (lcDoNotSuggestWords) via shared FilterSuggestions path.
		base.FilterNoSuggestWordsFn = base.filterNoSuggestWords
		// Java SpellingCheckRule: wordsToBeIgnored.addAll(userConfig.getAcceptedWords()) always.
		base.ApplyUserAcceptedWords(accepted)
	}
	base.IgnoreTokenFn = base.ruIgnoreToken
	// Java MorfologikSpellerRule.initSpeller for YO binary dict (+ spelling.txt etc.).
	base.InitSpellersFromGetters(nil, userWords)
	// Java YO ctor: same demos as base speller (каждя → каждая)
	base.AddExamplePair(
		rules.Wrong("Все счастливые семьи похожи друг на друга, <marker>каждя</marker> несчастливая семья несчастлива по-своему."),
		rules.Fixed("Все счастливые семьи похожи друг на друга, <marker>каждая</marker> несчастливая семья несчастлива по-своему."),
	)
	r := &MorfologikRussianYOSpellerRule{MorfologikRussianSpellerRule: base}
	// Java: setDefaultOff();
	r.SetDefaultOff()
	return r
}

// SetDefaultOff ports Rule.setDefaultOff().
func (r *MorfologikRussianYOSpellerRule) SetDefaultOff() {
	if r != nil {
		r.DefaultOff = true
	}
}

// IsDefaultOff ports Rule.isDefaultOff().
func (r *MorfologikRussianYOSpellerRule) IsDefaultOff() bool {
	return r != nil && r.DefaultOff
}

// GetDescription ports getDescription (ё-only experimental rule).
func (r *MorfologikRussianYOSpellerRule) GetDescription() string {
	return "Проверка орфографии. Только «Ё» (экспериментальное правило)."
}

// GetConfigureText ports MorfologikRussianYOSpellerRule.getConfigureText (same as base).
func (r *MorfologikRussianYOSpellerRule) GetConfigureText() string {
	return "Проверять слова на латинице, только термины (0/1)"
}

// GetRuleOptions ports MorfologikRussianYOSpellerRule.getRuleOptions.
// Java: new RuleOption(DEFAULT_MIN_RU_VALUE, getConfigureText(), 0, 1)
func (r *MorfologikRussianYOSpellerRule) GetRuleOptions() []*rules.RuleOption {
	cfg := "Проверять слова на латинице, только термины (0/1)"
	if r != nil {
		cfg = r.GetConfigureText()
	}
	return []*rules.RuleOption{
		rules.NewRuleOption(defaultMinRUValue, cfg, 0, 1),
	}
}
