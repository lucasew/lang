package ru

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikRussianYOSpellerRuleID = "MORFOLOGIK_RULE_RU_YO"
	RussianYOSpellerDict             = "/ru/hunspell/ru_RU_yo.dict"
)

// MorfologikRussianYOSpellerRule ports the ё-aware Russian speller.
type MorfologikRussianYOSpellerRule struct {
	*MorfologikRussianSpellerRule
}

func NewMorfologikRussianYOSpellerRule() *MorfologikRussianYOSpellerRule {
	base := NewMorfologikRussianSpellerRule()
	base.MorfologikSpellerRule = morfologik.NewMorfologikSpellerRule(
		MorfologikRussianYOSpellerRuleID, "ru", RussianYOSpellerDict, nil)
	return &MorfologikRussianYOSpellerRule{MorfologikRussianSpellerRule: base}
}
