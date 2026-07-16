package ru

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikRussianSpellerRuleID = "MORFOLOGIK_RULE_RU"
	RussianSpellerDict             = "/ru/hunspell/ru_RU.dict"
)

// MorfologikRussianSpellerRule ports rules.ru.MorfologikRussianSpellerRule.
type MorfologikRussianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikRussianSpellerRule() *MorfologikRussianSpellerRule {
	return &MorfologikRussianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikRussianSpellerRuleID, "ru", RussianSpellerDict, nil),
	}
}
